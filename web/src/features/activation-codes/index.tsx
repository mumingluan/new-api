/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Copy, Download, Filter, Plus, Settings2, Trash2 } from 'lucide-react'
import { type FormEvent, type ReactNode, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { CopyButton } from '@/components/copy-button'
import { DateTimePicker } from '@/components/datetime-picker'
import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { CompactDateTimeRangePicker } from '@/features/usage-logs/components/compact-date-time-range-picker'
import { formatTimestampToDate } from '@/lib/format'
import { useAuthStore } from '@/stores/auth-store'

import {
  createActivationCodes,
  deleteActivationCodes,
  exportActivationCodes,
  getActivationCodes,
  getActivationLogs,
  updateActivationCodes,
} from './api'
import type { ActivationCode, ActivationFilters, ActivationLog } from './types'

const EMPTY_FILTERS: ActivationFilters = {
  search: '',
  channel: '',
  status: '',
  days: '',
  createdFrom: '',
  createdTo: '',
}

const PAGE_SIZE = 20

function statusBadge(code: ActivationCode, t: (key: string) => string) {
  if (code.status === 2) {
    return <Badge variant='secondary'>{t('Used')}</Badge>
  }
  if (code.status === 3) {
    return <Badge variant='outline'>{t('Disabled')}</Badge>
  }
  if (code.expired_time < Math.floor(Date.now() / 1000)) {
    return <Badge variant='destructive'>{t('Expired')}</Badge>
  }
  return <Badge>{t('Active')}</Badge>
}

function CodesTable({
  codes,
  selected,
  onSelectionChange,
}: {
  codes: ActivationCode[]
  selected: Set<number>
  onSelectionChange: (selected: Set<number>) => void
}) {
  const { t } = useTranslation()
  const allSelected =
    codes.length > 0 && codes.every((code) => selected.has(code.id))

  const toggle = (id: number, checked: boolean) => {
    const next = new Set(selected)
    if (checked) next.add(id)
    else next.delete(id)
    onSelectionChange(next)
  }

  return (
    <>
      <div className='hidden overflow-x-auto md:block'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-10'>
                <Checkbox
                  aria-label={t('Select all')}
                  checked={allSelected}
                  indeterminate={
                    !allSelected && codes.some((code) => selected.has(code.id))
                  }
                  onCheckedChange={(checked) =>
                    onSelectionChange(
                      checked
                        ? new Set(codes.map((code) => code.id))
                        : new Set()
                    )
                  }
                />
              </TableHead>
              <TableHead>{t('Activation code')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Channel')}</TableHead>
              <TableHead>{t('Duration')}</TableHead>
              <TableHead>{t('Expires')}</TableHead>
              <TableHead>{t('Created')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {codes.map((code) => (
              <TableRow
                key={code.id}
                data-state={selected.has(code.id) ? 'selected' : undefined}
              >
                <TableCell>
                  <Checkbox
                    aria-label={t('Select row')}
                    checked={selected.has(code.id)}
                    onCheckedChange={(checked) => toggle(code.id, !!checked)}
                  />
                </TableCell>
                <TableCell>
                  <div className='flex min-w-64 items-center gap-1'>
                    <code className='text-xs break-all'>{code.code}</code>
                    <CopyButton value={code.code} />
                  </div>
                </TableCell>
                <TableCell>{statusBadge(code, t)}</TableCell>
                <TableCell>{code.channel}</TableCell>
                <TableCell>
                  {t('{{count}} days', { count: code.days })}
                </TableCell>
                <TableCell className='whitespace-nowrap'>
                  {formatTimestampToDate(code.expired_time)}
                </TableCell>
                <TableCell className='whitespace-nowrap'>
                  {formatTimestampToDate(code.created_time)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className='grid gap-3 md:hidden'>
        {codes.map((code) => (
          <Card key={code.id} size='sm'>
            <CardHeader className='flex-row items-start gap-3'>
              <Checkbox
                aria-label={t('Select row')}
                checked={selected.has(code.id)}
                onCheckedChange={(checked) => toggle(code.id, !!checked)}
              />
              <div className='min-w-0 flex-1'>
                <CardTitle className='flex items-start justify-between gap-2 text-sm'>
                  <code className='break-all'>{code.code}</code>
                  <CopyButton value={code.code} />
                </CardTitle>
                <div className='mt-2 flex flex-wrap gap-2'>
                  {statusBadge(code, t)}
                  <Badge variant='outline'>{code.channel}</Badge>
                  <Badge variant='outline'>
                    {t('{{count}} days', { count: code.days })}
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent className='text-muted-foreground grid gap-1 text-xs'>
              <span>
                {t('Expires')}: {formatTimestampToDate(code.expired_time)}
              </span>
              <span>
                {t('Created')}: {formatTimestampToDate(code.created_time)}
              </span>
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  )
}

function CreateDialog({
  open,
  onOpenChange,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: (codes: ActivationCode[]) => void
}) {
  const { t } = useTranslation()
  const userId = useAuthStore((state) => state.auth.user?.id ?? 0)
  const [count, setCount] = useState(1)
  const [days, setDays] = useState(30)
  const [channel, setChannel] = useState('default')
  const [expiry, setExpiry] = useState<Date | undefined>(
    () => new Date(Date.now() + 365 * 86400000)
  )
  const [customCodes, setCustomCodes] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const codes = customCodes
      .split(/\r?\n/)
      .map((code) => code.trim())
      .filter(Boolean)
    if (!expiry) {
      toast.error(t('Expiration date'))
      return
    }
    const expiredTime = Math.floor(expiry.getTime() / 1000)
    setSubmitting(true)
    try {
      const result = await createActivationCodes({
        count: codes.length || count,
        days,
        channel: channel.trim(),
        expired_time: expiredTime,
        codes,
      })
      onSuccess(result.data)
      onOpenChange(false)
      toast.success(
        t('Created {{count}} activation codes.', { count: result.data.length })
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-xl'>
        <DialogHeader>
          <DialogTitle>{t('Batch create activation codes')}</DialogTitle>
          <DialogDescription>
            {t('New codes always start with your user ID: {{prefix}}', {
              prefix: `${userId}_`,
            })}
          </DialogDescription>
        </DialogHeader>
        <form id='activation-create-form' onSubmit={submit}>
          <FieldGroup>
            <div className='grid gap-4 sm:grid-cols-2'>
              <Field>
                <FieldLabel htmlFor='activation-count'>
                  {t('Quantity')}
                </FieldLabel>
                <Input
                  id='activation-count'
                  type='number'
                  min={1}
                  max={1000}
                  value={count}
                  onChange={(event) => setCount(Number(event.target.value))}
                  disabled={customCodes.trim().length > 0}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor='activation-days'>
                  {t('Duration (days)')}
                </FieldLabel>
                <Input
                  id='activation-days'
                  type='number'
                  min={1}
                  value={days}
                  onChange={(event) => setDays(Number(event.target.value))}
                />
              </Field>
            </div>
            <Field>
              <FieldLabel htmlFor='activation-channel'>
                {t('Channel')}
              </FieldLabel>
              <Input
                id='activation-channel'
                value={channel}
                maxLength={100}
                onChange={(event) => setChannel(event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='activation-expiry'>
                {t('Expiration date')}
              </FieldLabel>
              <DateTimePicker
                value={expiry}
                onChange={setExpiry}
                className='w-full'
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='activation-custom-codes'>
                {t('Custom code parts')}
              </FieldLabel>
              <Textarea
                id='activation-custom-codes'
                value={customCodes}
                onChange={(event) => setCustomCodes(event.target.value)}
                placeholder={'TEST0001\nTEST0002'}
              />
              <FieldDescription>
                {t(
                  'Enter one code part per line. The user ID prefix is added automatically.'
                )}
              </FieldDescription>
            </Field>
          </FieldGroup>
        </form>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            {t('Cancel')}
          </Button>
          <Button
            form='activation-create-form'
            type='submit'
            disabled={submitting}
          >
            {submitting ? t('Creating...') : t('Create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function BatchManageDialog({
  open,
  onOpenChange,
  ids,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  ids: number[]
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const [days, setDays] = useState('')
  const [channel, setChannel] = useState('')
  const [expiry, setExpiry] = useState<Date | undefined>()
  const [status, setStatus] = useState('unchanged')
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    try {
      await updateActivationCodes({
        ids,
        days: days ? Number(days) : undefined,
        channel: channel.trim() || undefined,
        expired_time: expiry ? Math.floor(expiry.getTime() / 1000) : undefined,
        status: status === 'unchanged' ? undefined : Number(status),
      })
      toast.success(t('Selected activation codes updated.'))
      onSuccess()
      onOpenChange(false)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Batch manage activation codes')}</DialogTitle>
          <DialogDescription>
            {t(
              'Only filled fields will be applied to {{count}} selected codes.',
              {
                count: ids.length,
              }
            )}
          </DialogDescription>
        </DialogHeader>
        <form id='activation-batch-form' onSubmit={submit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor='batch-days'>
                {t('Duration (days)')}
              </FieldLabel>
              <Input
                id='batch-days'
                type='number'
                min={1}
                value={days}
                onChange={(event) => setDays(event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='batch-channel'>{t('Channel')}</FieldLabel>
              <Input
                id='batch-channel'
                value={channel}
                onChange={(event) => setChannel(event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='batch-expiry'>
                {t('Expiration date')}
              </FieldLabel>
              <DateTimePicker
                value={expiry}
                onChange={setExpiry}
                className='w-full'
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='batch-status'>{t('Status')}</FieldLabel>
              <Select
                value={status}
                onValueChange={(value) => setStatus(value ?? 'unchanged')}
                items={[
                  { value: 'unchanged', label: t('Keep unchanged') },
                  { value: '1', label: t('Active') },
                  { value: '3', label: t('Disabled') },
                ]}
              >
                <SelectTrigger id='batch-status' className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  <SelectGroup>
                    <SelectItem value='unchanged'>
                      {t('Keep unchanged')}
                    </SelectItem>
                    <SelectItem value='1'>{t('Active')}</SelectItem>
                    <SelectItem value='3'>{t('Disabled')}</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
        </form>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            {t('Cancel')}
          </Button>
          <Button
            form='activation-batch-form'
            type='submit'
            disabled={submitting}
          >
            {submitting ? t('Saving...') : t('Save changes')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function DeleteByCodeDialog({
  open,
  onOpenChange,
  onSuccess,
  initialCodes = [],
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  initialCodes?: string[]
}) {
  const { t } = useTranslation()
  const [value, setValue] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const effectiveValue = value || initialCodes.join('\n')

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const codes = effectiveValue
      .split(/[\r\n,]+/)
      .map((code) => code.trim())
      .filter(Boolean)
    if (codes.length === 0) {
      toast.error(t('Enter at least one activation code.'))
      return
    }
    setSubmitting(true)
    try {
      const result = await deleteActivationCodes({ codes })
      toast.success(
        t('Deleted {{count}} activation codes.', {
          count: result.data.deleted,
        })
      )
      onSuccess()
      onOpenChange(false)
      setValue('')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Delete activation codes by code')}</DialogTitle>
          <DialogDescription>
            {t('Used activation codes are retained with their usage records.')}
          </DialogDescription>
        </DialogHeader>
        <form id='activation-delete-form' onSubmit={submit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor='delete-codes'>
                {t('Activation codes')}
              </FieldLabel>
              <Textarea
                id='delete-codes'
                value={effectiveValue}
                onChange={(event) => setValue(event.target.value)}
              />
              <FieldDescription>
                {t('Enter one code per line or separate codes with commas.')}
              </FieldDescription>
            </Field>
          </FieldGroup>
        </form>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            {t('Cancel')}
          </Button>
          <Button
            variant='destructive'
            form='activation-delete-form'
            type='submit'
            disabled={submitting}
          >
            {submitting ? t('Deleting...') : t('Delete')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function Pagination({
  page,
  total,
  onPageChange,
}: {
  page: number
  total: number
  onPageChange: (page: number) => void
}) {
  const { t } = useTranslation()
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  return (
    <div className='flex items-center justify-between gap-3 border-t pt-4'>
      <span className='text-muted-foreground text-sm'>
        {t('{{total}} records', { total })}
      </span>
      <div className='flex items-center gap-2'>
        <Button
          variant='outline'
          size='sm'
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          {t('Previous')}
        </Button>
        <span className='text-sm tabular-nums'>
          {page} / {pages}
        </span>
        <Button
          variant='outline'
          size='sm'
          disabled={page >= pages}
          onClick={() => onPageChange(page + 1)}
        >
          {t('Next')}
        </Button>
      </div>
    </div>
  )
}

function UsageLogs() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [action, setAction] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['activation-logs', page, search, action],
    queryFn: () => getActivationLogs(page, PAGE_SIZE, search, action),
    placeholderData: (previous) => previous,
  })
  const logs = data?.data.items ?? []
  let logsContent: ReactNode
  if (isLoading) {
    logsContent = (
      <p className='text-muted-foreground py-12 text-center'>
        {t('Loading...')}
      </p>
    )
  } else if (logs.length === 0) {
    logsContent = (
      <p className='text-muted-foreground py-12 text-center'>
        {t('No activation records')}
      </p>
    )
  } else {
    logsContent = (
      <>
        <div className='hidden overflow-x-auto md:block'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Activation code')}</TableHead>
                <TableHead>{t('Operation')}</TableHead>
                <TableHead>{t('Identifier')}</TableHead>
                <TableHead>{t('API key')}</TableHead>
                <TableHead>{t('IP address')}</TableHead>
                <TableHead>{t('Used at')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log: ActivationLog) => (
                <TableRow key={log.id}>
                  <TableCell>
                    <code className='text-xs'>{log.activation_code}</code>
                  </TableCell>
                  <TableCell>
                    <Badge variant='outline'>
                      {log.action === 'create' ? t('Create') : t('Renew key')}
                    </Badge>
                  </TableCell>
                  <TableCell>{log.identifier}</TableCell>
                  <TableCell>
                    <div className='flex min-w-56 items-center gap-1'>
                      <code className='truncate text-xs'>{log.api_key}</code>
                      <CopyButton value={log.api_key} />
                    </div>
                  </TableCell>
                  <TableCell>{log.client_ip}</TableCell>
                  <TableCell className='whitespace-nowrap'>
                    {formatTimestampToDate(log.used_time)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <div className='grid gap-3 md:hidden'>
          {logs.map((log: ActivationLog) => (
            <Card key={log.id} size='sm'>
              <CardHeader>
                <CardTitle className='flex items-start justify-between gap-2 text-sm'>
                  <code className='break-all'>{log.activation_code}</code>
                  <Badge variant='outline'>
                    {log.action === 'create' ? t('Create') : t('Renew key')}
                  </Badge>
                </CardTitle>
              </CardHeader>
              <CardContent className='grid gap-2 text-sm'>
                <span className='text-muted-foreground'>{log.identifier}</span>
                <div className='flex items-center gap-1'>
                  <code className='min-w-0 flex-1 text-xs break-all'>
                    {log.api_key}
                  </code>
                  <CopyButton value={log.api_key} />
                </div>
                <span className='text-muted-foreground text-xs'>
                  {formatTimestampToDate(log.used_time)} · {log.client_ip}
                </span>
              </CardContent>
            </Card>
          ))}
        </div>
      </>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Activation code usage records')}</CardTitle>
      </CardHeader>
      <CardContent className='grid gap-4'>
        <div className='grid gap-3 sm:grid-cols-[1fr_180px]'>
          <Input
            value={search}
            onChange={(event) => {
              setSearch(event.target.value)
              setPage(1)
            }}
            placeholder={t('Filter by activation code or identifier...')}
          />
          <Select
            value={action || 'all'}
            onValueChange={(value) => {
              setAction(value === 'all' ? '' : (value ?? ''))
              setPage(1)
            }}
            items={[
              { value: 'all', label: t('All operations') },
              { value: 'create', label: t('Create key') },
              { value: 'renew', label: t('Renew key') },
            ]}
          >
            <SelectTrigger className='w-full'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent alignItemWithTrigger={false}>
              <SelectGroup>
                <SelectItem value='all'>{t('All operations')}</SelectItem>
                <SelectItem value='create'>{t('Create key')}</SelectItem>
                <SelectItem value='renew'>{t('Renew key')}</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        {logsContent}
        <Pagination
          page={page}
          total={data?.data.total ?? 0}
          onPageChange={setPage}
        />
      </CardContent>
    </Card>
  )
}

export function ActivationCodes() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<ActivationFilters>(EMPTY_FILTERS)
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [createOpen, setCreateOpen] = useState(false)
  const [manageOpen, setManageOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createdCodes, setCreatedCodes] = useState<ActivationCode[]>([])
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['activation-codes', page, filters],
    queryFn: () => getActivationCodes(page, PAGE_SIZE, filters),
    placeholderData: (previous) => previous,
  })
  const codes = data?.data.items ?? []
  const selectedCodes = codes.filter((code) => selected.has(code.id))
  let codesContent: ReactNode
  if (isLoading) {
    codesContent = (
      <p className='text-muted-foreground py-16 text-center'>
        {t('Loading...')}
      </p>
    )
  } else if (codes.length === 0) {
    codesContent = (
      <p className='text-muted-foreground py-16 text-center'>
        {t('No activation codes found')}
      </p>
    )
  } else {
    codesContent = (
      <CodesTable
        codes={codes}
        selected={selected}
        onSelectionChange={setSelected}
      />
    )
  }

  const refresh = () => {
    setSelected(new Set())
    void queryClient.invalidateQueries({ queryKey: ['activation-codes'] })
    void queryClient.invalidateQueries({ queryKey: ['activation-logs'] })
  }

  return (
    <>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>
          {t('Activation Code Management')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <Button
            variant='outline'
            onClick={() => void exportActivationCodes(filters)}
          >
            <Download />
            {t('Export CSV')}
          </Button>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            {t('Batch create')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='h-full min-h-0 overflow-y-auto pr-1'>
            <Tabs
              defaultValue='codes'
              className='grid min-h-full content-start gap-4'
            >
              <TabsList className='w-fit'>
                <TabsTrigger value='codes'>{t('Activation codes')}</TabsTrigger>
                <TabsTrigger value='logs'>{t('Usage records')}</TabsTrigger>
              </TabsList>
              <TabsContent value='codes' className='min-h-0'>
                <Card>
                  <CardHeader className='gap-4'>
                    <CardTitle className='flex items-center gap-2 text-base'>
                      <Filter className='size-4' />
                      {t('Advanced filters')}
                    </CardTitle>
                    <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-6'>
                      <Input
                        className='lg:col-span-2'
                        value={filters.search}
                        onChange={(event) => {
                          setFilters((current) => ({
                            ...current,
                            search: event.target.value,
                          }))
                          setPage(1)
                        }}
                        placeholder={t('Search activation codes...')}
                      />
                      <Input
                        value={filters.channel}
                        onChange={(event) => {
                          setFilters((current) => ({
                            ...current,
                            channel: event.target.value,
                          }))
                          setPage(1)
                        }}
                        placeholder={t('Channel')}
                      />
                      <Input
                        type='number'
                        min={1}
                        value={filters.days}
                        onChange={(event) => {
                          setFilters((current) => ({
                            ...current,
                            days: event.target.value,
                          }))
                          setPage(1)
                        }}
                        placeholder={t('Duration (days)')}
                      />
                      <Select
                        value={filters.status || 'all'}
                        onValueChange={(value) => {
                          setFilters((current) => ({
                            ...current,
                            status: value === 'all' ? '' : (value ?? ''),
                          }))
                          setPage(1)
                        }}
                        items={[
                          { value: 'all', label: t('All statuses') },
                          { value: '1', label: t('Active') },
                          { value: '2', label: t('Used') },
                          { value: '3', label: t('Disabled') },
                        ]}
                      >
                        <SelectTrigger className='w-full'>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            <SelectItem value='all'>
                              {t('All statuses')}
                            </SelectItem>
                            <SelectItem value='1'>{t('Active')}</SelectItem>
                            <SelectItem value='2'>{t('Used')}</SelectItem>
                            <SelectItem value='3'>{t('Disabled')}</SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                      <Button
                        variant='ghost'
                        onClick={() => {
                          setFilters(EMPTY_FILTERS)
                          setPage(1)
                        }}
                      >
                        {t('Reset filters')}
                      </Button>
                      <div className='sm:col-span-2 lg:col-span-3'>
                        <CompactDateTimeRangePicker
                          start={
                            filters.createdFrom
                              ? new Date(filters.createdFrom)
                              : undefined
                          }
                          end={
                            filters.createdTo
                              ? new Date(filters.createdTo)
                              : undefined
                          }
                          onChange={({ start, end }) => {
                            setFilters((current) => ({
                              ...current,
                              createdFrom: start?.toISOString() ?? '',
                              createdTo: end?.toISOString() ?? '',
                            }))
                            setPage(1)
                          }}
                        />
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className='grid min-h-0 gap-4'>
                    {selected.size > 0 ? (
                      <Alert>
                        <AlertDescription className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
                          <span>
                            {t('{{count}} activation codes selected.', {
                              count: selected.size,
                            })}
                          </span>
                          <div className='flex flex-wrap gap-2'>
                            <CopyButton
                              value={selectedCodes
                                .map((code) => code.code)
                                .join('\n')}
                              variant='outline'
                              size='sm'
                            >
                              <Copy />
                              {t('Batch copy')}
                            </CopyButton>
                            <Button
                              variant='outline'
                              size='sm'
                              onClick={() => setManageOpen(true)}
                            >
                              <Settings2 />
                              {t('Batch manage')}
                            </Button>
                            <Button
                              variant='destructive'
                              size='sm'
                              onClick={() => setDeleteOpen(true)}
                            >
                              <Trash2 />
                              {t('Delete by code')}
                            </Button>
                          </div>
                        </AlertDescription>
                      </Alert>
                    ) : null}
                    {createdCodes.length > 0 ? (
                      <Alert>
                        <AlertDescription className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
                          <span>
                            {t('{{count}} newly created codes are ready.', {
                              count: createdCodes.length,
                            })}
                          </span>
                          <CopyButton
                            value={createdCodes
                              .map((code) => code.code)
                              .join('\n')}
                            variant='outline'
                            size='sm'
                          >
                            <Copy />
                            {t('Copy all')}
                          </CopyButton>
                        </AlertDescription>
                      </Alert>
                    ) : null}
                    {codesContent}
                    {isFetching && !isLoading ? (
                      <span className='text-muted-foreground text-xs'>
                        {t('Refreshing...')}
                      </span>
                    ) : null}
                    <Pagination
                      page={page}
                      total={data?.data.total ?? 0}
                      onPageChange={(next) => {
                        setPage(next)
                        setSelected(new Set())
                      }}
                    />
                  </CardContent>
                </Card>
              </TabsContent>
              <TabsContent value='logs'>
                <UsageLogs />
              </TabsContent>
            </Tabs>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
      <CreateDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSuccess={(newCodes) => {
          setCreatedCodes(newCodes)
          refresh()
        }}
      />
      <BatchManageDialog
        open={manageOpen}
        onOpenChange={setManageOpen}
        ids={[...selected]}
        onSuccess={refresh}
      />
      <DeleteByCodeDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        initialCodes={selectedCodes.map((code) => code.code)}
        onSuccess={refresh}
      />
    </>
  )
}
