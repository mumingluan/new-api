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
import { useMutation } from '@tanstack/react-query'
import {
  BarChart3,
  Clock3,
  Download,
  Eye,
  KeyRound,
  Play,
  ShieldAlert,
  Users,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CompactDateTimeRangePicker } from '@/features/usage-logs/components/compact-date-time-range-picker'
import dayjs from '@/lib/dayjs'
import { formatQuota, parseQuotaFromDollars } from '@/lib/format'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import {
  executeKeyBatch,
  exportKeyStatistics,
  getKeyStatistics,
  previewKeyBatch,
} from './api'
import type {
  KeyBatchAction,
  KeyBatchOperationPayload,
  KeyBatchPreview,
  KeyStatisticsGroup,
  KeyStatisticsQuery,
  KeyStatisticsResult,
  KeyStatisticsSort,
} from './types'

const numberFormatter = new Intl.NumberFormat()
const parseNonNegative = (value: string) => {
  const number = Number(value)
  return Number.isFinite(number) && number >= 0 ? number : 0
}
const parseOptionalPositiveInt = (value: string) => {
  const number = Number.parseInt(value, 10)
  return Number.isFinite(number) && number > 0 ? number : undefined
}

function ScopeSwitch({
  checked,
  onCheckedChange,
  isAdmin,
}: {
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  isAdmin: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className='bg-muted/25 flex items-start justify-between gap-4 rounded-lg border p-3'>
      <div className='min-w-0'>
        <div className='flex items-center gap-2 text-sm font-medium'>
          <Users className='size-4' />
          {t('Operate on all users')}
        </div>
        <p className='text-muted-foreground mt-1 text-xs'>
          {isAdmin
            ? t(
                'Off by default. Enable only when this query or operation should include every user key.'
              )
            : t('Only administrators can include keys owned by other users.')}
        </p>
      </div>
      <Switch
        checked={checked}
        onCheckedChange={onCheckedChange}
        disabled={!isAdmin}
        aria-label={t('Operate on all users')}
      />
    </div>
  )
}

function PreviewCards({ preview }: { preview: KeyBatchPreview }) {
  const { t } = useTranslation()
  const cards = [
    [t('Matched keys'), preview.matched_tokens],
    [t('Actionable keys'), preview.actionable_tokens],
    [t('Affected users'), preview.affected_users],
    [t('Used keys'), preview.used_tokens],
  ] as const
  return (
    <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
      {cards.map(([label, value]) => (
        <Card key={label} data-card-hover='false' className='gap-2 py-4'>
          <CardHeader className='px-4 pb-0'>
            <CardDescription>{label}</CardDescription>
            <CardTitle className='text-2xl tabular-nums'>
              {numberFormatter.format(value)}
            </CardTitle>
          </CardHeader>
        </Card>
      ))}
    </div>
  )
}

function BatchOperationsTab({ isAdmin }: { isAdmin: boolean }) {
  const { t } = useTranslation()
  const [action, setAction] = useState<KeyBatchAction>('extend_expiry')
  const [allUsers, setAllUsers] = useState(false)
  const [usedOnly, setUsedOnly] = useState(false)
  const [minQuotaEnabled, setMinQuotaEnabled] = useState(false)
  const [minQuota, setMinQuota] = useState('')
  const [days, setDays] = useState('30')
  const [hours, setHours] = useState('0')
  const [minutes, setMinutes] = useState('0')
  const [quotaAmount, setQuotaAmount] = useState('1')
  const [preview, setPreview] = useState<KeyBatchPreview | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const isTimeAction = action === 'extend_expiry' || action === 'deduct_expiry'

  const payload = useMemo<KeyBatchOperationPayload>(
    () => ({
      action,
      duration_seconds: Math.round(
        parseNonNegative(days) * 86400 +
          parseNonNegative(hours) * 3600 +
          parseNonNegative(minutes) * 60
      ),
      quota: Math.max(0, parseQuotaFromDollars(parseNonNegative(quotaAmount))),
      filter: {
        all_users: allUsers,
        used_only: usedOnly,
        min_remaining_quota: minQuotaEnabled
          ? Math.max(0, parseQuotaFromDollars(parseNonNegative(minQuota)))
          : undefined,
      },
    }),
    [
      action,
      allUsers,
      days,
      hours,
      minQuota,
      minQuotaEnabled,
      minutes,
      quotaAmount,
      usedOnly,
    ]
  )

  const validate = () => {
    if (isTimeAction && payload.duration_seconds <= 0) {
      toast.error(t('Enter a duration greater than zero.'))
      return false
    }
    if (!isTimeAction && payload.quota <= 0) {
      toast.error(t('Enter a quota amount greater than zero.'))
      return false
    }
    return true
  }
  const previewMutation = useMutation({
    mutationFn: previewKeyBatch,
    onSuccess: setPreview,
    onError: () => toast.error(t('Preview failed')),
  })
  const executeMutation = useMutation({
    mutationFn: executeKeyBatch,
    onSuccess: async (data) => {
      toast.success(
        t('{{count}} keys updated successfully.', { count: data.affected })
      )
      setConfirmOpen(false)
      setPreview(await previewKeyBatch(payload))
    },
    onError: () => toast.error(t('Batch operation failed')),
  })
  const actions = [
    { value: 'extend_expiry', label: t('Extend key expiration') },
    { value: 'add_quota', label: t('Add key quota') },
    { value: 'deduct_expiry', label: t('Deduct key expiration') },
    { value: 'deduct_quota', label: t('Deduct key quota') },
  ]

  return (
    <div className='min-w-0 max-w-full space-y-4'>
      <Card data-card-hover='false' className='min-w-0 max-w-full'>
        <CardHeader>
          <CardTitle className='flex items-center gap-2 text-base'>
            <KeyRound className='size-4' />
            {t('Batch operation settings')}
          </CardTitle>
          <CardDescription>
            {t(
              'The default scope is always your own keys. Permanent keys ignore expiration changes, and unlimited keys ignore quota changes.'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>
          <ScopeSwitch
            checked={allUsers}
            onCheckedChange={(checked) => {
              setAllUsers(checked)
              setPreview(null)
            }}
            isAdmin={isAdmin}
          />
          {allUsers ? (
            <Alert variant='destructive'>
              <ShieldAlert />
              <AlertTitle>{t('All-user scope enabled')}</AlertTitle>
              <AlertDescription>
                {t(
                  'This operation can change keys owned by every user who matches the filters.'
                )}
              </AlertDescription>
            </Alert>
          ) : null}
          <div className='grid gap-4 lg:grid-cols-2'>
            <Field>
              <FieldLabel>{t('Operation')}</FieldLabel>
              <Select
                value={action}
                onValueChange={(value) => {
                  setAction((value ?? 'extend_expiry') as KeyBatchAction)
                  setPreview(null)
                }}
                items={actions}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {actions.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            {isTimeAction ? (
              <Field>
                <FieldLabel>{t('Duration')}</FieldLabel>
                <div className='grid grid-cols-3 gap-2'>
                  <Input
                    type='number'
                    min={0}
                    value={days}
                    onChange={(event) => setDays(event.target.value)}
                    placeholder={t('Days')}
                  />
                  <Input
                    type='number'
                    min={0}
                    max={23}
                    value={hours}
                    onChange={(event) => setHours(event.target.value)}
                    placeholder={t('Hours')}
                  />
                  <Input
                    type='number'
                    min={0}
                    max={59}
                    value={minutes}
                    onChange={(event) => setMinutes(event.target.value)}
                    placeholder={t('Minutes')}
                  />
                </div>
              </Field>
            ) : (
              <Field>
                <FieldLabel>{t('Quota amount')}</FieldLabel>
                <Input
                  type='number'
                  min={0}
                  step='any'
                  value={quotaAmount}
                  onChange={(event) => setQuotaAmount(event.target.value)}
                />
                <FieldDescription>
                  {t(
                    'Uses the quota or currency unit configured by the system.'
                  )}
                </FieldDescription>
              </Field>
            )}
          </div>
          <div className='grid gap-3 md:grid-cols-2'>
            <label className='flex items-start gap-3 rounded-lg border p-3'>
              <Checkbox
                checked={usedOnly}
                onCheckedChange={(checked) => {
                  setUsedOnly(checked === true)
                  setPreview(null)
                }}
              />
              <span>
                <span className='block text-sm font-medium'>
                  {t('Used keys only')}
                </span>
                <span className='text-muted-foreground block text-xs'>
                  {t(
                    'Only include keys whose used quota is greater than zero.'
                  )}
                </span>
              </span>
            </label>
            <div className='rounded-lg border p-3'>
              <div className='flex items-center justify-between gap-3'>
                <div>
                  <p className='text-sm font-medium'>
                    {t('Minimum remaining quota')}
                  </p>
                  <p className='text-muted-foreground text-xs'>
                    {t('Only include keys above this remaining quota.')}
                  </p>
                </div>
                <Switch
                  checked={minQuotaEnabled}
                  onCheckedChange={(checked) => {
                    setMinQuotaEnabled(checked)
                    setPreview(null)
                  }}
                />
              </div>
              {minQuotaEnabled ? (
                <Input
                  className='mt-3'
                  type='number'
                  min={0}
                  step='any'
                  value={minQuota}
                  onChange={(event) => setMinQuota(event.target.value)}
                  placeholder='0'
                />
              ) : null}
            </div>
          </div>
          <div className='flex flex-col-reverse gap-2 border-t pt-4 sm:flex-row sm:justify-end'>
            <Button
              variant='outline'
              onClick={() => {
                if (validate()) previewMutation.mutate(payload)
              }}
              disabled={previewMutation.isPending}
            >
              <Eye />
              {previewMutation.isPending
                ? t('Loading...')
                : t('Preview impact')}
            </Button>
            <Button
              onClick={() => {
                if (validate()) setConfirmOpen(true)
              }}
              disabled={executeMutation.isPending}
            >
              <Play />
              {t('Execute batch operation')}
            </Button>
          </div>
        </CardContent>
      </Card>
      {preview ? (
        <>
          <PreviewCards preview={preview} />
          <Alert>
            <AlertDescription>
              {t(
                'Permanent keys: {{permanent}}; unlimited keys: {{unlimited}}.',
                {
                  permanent: preview.permanent_tokens,
                  unlimited: preview.unlimited_tokens,
                }
              )}{' '}
              {t('Remaining quota: {{remaining}}; used quota: {{used}}.', {
                remaining: formatQuota(preview.total_remaining_quota),
                used: formatQuota(preview.total_used_quota),
              })}
            </AlertDescription>
          </Alert>
        </>
      ) : null}
      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Confirm batch operation')}</DialogTitle>
            <DialogDescription>
              {allUsers
                ? t(
                    'All-user scope is enabled. This action cannot be automatically undone.'
                  )
                : t(
                    'This action applies only to your matching keys and cannot be automatically undone.'
                  )}
            </DialogDescription>
          </DialogHeader>
          <div className='bg-muted/30 space-y-2 rounded-lg border p-3 text-sm'>
            <div className='flex justify-between gap-4'>
              <span className='text-muted-foreground'>{t('Operation')}</span>
              <span>
                {actions.find((item) => item.value === action)?.label}
              </span>
            </div>
            {preview ? (
              <div className='flex justify-between gap-4'>
                <span className='text-muted-foreground'>
                  {t('Actionable keys')}
                </span>
                <span className='tabular-nums'>
                  {preview.actionable_tokens}
                </span>
              </div>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setConfirmOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button
              variant={allUsers ? 'destructive' : 'default'}
              onClick={() => executeMutation.mutate(payload)}
              disabled={executeMutation.isPending}
            >
              {executeMutation.isPending ? t('Processing...') : t('Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function StatisticsTotals({ data }: { data: KeyStatisticsResult }) {
  const { t } = useTranslation()
  const cards = [
    [t('Requests'), numberFormatter.format(data.totals.request_count)],
    [t('Input Tokens'), numberFormatter.format(data.totals.prompt_tokens)],
    [t('Output Tokens'), numberFormatter.format(data.totals.completion_tokens)],
    [t('Unique users'), numberFormatter.format(data.totals.unique_users)],
  ] as const
  return (
    <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
      {cards.map(([label, value]) => (
        <Card key={label} data-card-hover='false' className='gap-2 py-4'>
          <CardHeader className='px-4 pb-0'>
            <CardDescription>{label}</CardDescription>
            <CardTitle className='text-xl tabular-nums'>{value}</CardTitle>
          </CardHeader>
        </Card>
      ))}
    </div>
  )
}

function StatisticsTab({ isAdmin }: { isAdmin: boolean }) {
  const { t } = useTranslation()
  const [allUsers, setAllUsers] = useState(false)
  const [start, setStart] = useState(() =>
    dayjs().subtract(6, 'day').startOf('day').toDate()
  )
  const [end, setEnd] = useState(() => dayjs().endOf('day').toDate())
  const [groupBy, setGroupBy] = useState<KeyStatisticsGroup>('token_name')
  const [sortBy, setSortBy] = useState<KeyStatisticsSort>('request_count')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [userId, setUserId] = useState('')
  const [excludeUserId, setExcludeUserId] = useState('')
  const [model, setModel] = useState('')
  const [minTokens, setMinTokens] = useState('0')
  const [top, setTop] = useState('50')
  const [result, setResult] = useState<KeyStatisticsResult | null>(null)

  const query = useMemo<KeyStatisticsQuery>(
    () => ({
      all_users: allUsers,
      start_timestamp: Math.floor(start.getTime() / 1000),
      end_timestamp: Math.floor(end.getTime() / 1000) + 1,
      group_by: groupBy,
      sort_by: sortBy,
      sort_order: sortOrder,
      user_id: allUsers ? parseOptionalPositiveInt(userId) : undefined,
      exclude_user_id: allUsers
        ? parseOptionalPositiveInt(excludeUserId)
        : undefined,
      model: model.trim() || undefined,
      min_tokens: Math.round(parseNonNegative(minTokens)),
      top: Math.round(parseNonNegative(top)),
    }),
    [
      allUsers,
      end,
      excludeUserId,
      groupBy,
      minTokens,
      model,
      sortBy,
      sortOrder,
      start,
      top,
      userId,
    ]
  )

  const queryMutation = useMutation({
    mutationFn: getKeyStatistics,
    onSuccess: setResult,
    onError: () => toast.error(t('Statistics query failed')),
  })
  const exportMutation = useMutation({
    mutationFn: exportKeyStatistics,
    onSuccess: () => toast.success(t('CSV exported successfully')),
    onError: () => toast.error(t('CSV export failed')),
  })
  const validate = () => {
    if (end <= start) {
      toast.error(t('End time must be later than start time.'))
      return false
    }
    if (query.top < 1 || query.top > 500) {
      toast.error(t('Top N must be between 1 and 500.'))
      return false
    }
    return true
  }
  const groups = [
    { value: 'token_name', label: t('Key name') },
    { value: 'model_name', label: t('Model') },
    { value: 'username', label: t('Username') },
    { value: 'channel_name', label: t('Channel') },
    { value: 'user_id', label: t('User ID') },
  ]
  const sorts = [
    { value: 'request_count', label: t('Request count') },
    { value: 'prompt_tokens', label: t('Input Tokens') },
    { value: 'completion_tokens', label: t('Output Tokens') },
    { value: 'quota', label: t('Quota') },
  ]

  return (
    <div className='min-w-0 max-w-full space-y-4'>
      <Card data-card-hover='false' className='min-w-0 max-w-full'>
        <CardHeader>
          <CardTitle className='flex items-center gap-2 text-base'>
            <BarChart3 className='size-4' />
            {t('Usage log statistics')}
          </CardTitle>
          <CardDescription>
            {t(
              'Aggregate consume logs by key, model, user, or channel. Zero-sided token usage is included.'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>
          <ScopeSwitch
            checked={allUsers}
            onCheckedChange={(checked) => {
              setAllUsers(checked)
              setResult(null)
            }}
            isAdmin={isAdmin}
          />
          <div className='grid gap-4 lg:grid-cols-3'>
            <Field className='lg:col-span-2'>
              <FieldLabel>{t('Date Range')}</FieldLabel>
              <CompactDateTimeRangePicker
                start={start}
                end={end}
                onChange={(range) => {
                  if (range.start) setStart(range.start)
                  if (range.end) setEnd(range.end)
                }}
              />
            </Field>
            <Field>
              <FieldLabel>{t('Top N')}</FieldLabel>
              <Input
                type='number'
                min={1}
                max={500}
                value={top}
                onChange={(event) => setTop(event.target.value)}
              />
            </Field>
          </div>
          <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
            <Field>
              <FieldLabel>{t('Group by')}</FieldLabel>
              <Select
                value={groupBy}
                onValueChange={(value) =>
                  setGroupBy((value ?? 'token_name') as KeyStatisticsGroup)
                }
                items={groups}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {groups.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>{t('Sort by')}</FieldLabel>
              <Select
                value={sortBy}
                onValueChange={(value) =>
                  setSortBy((value ?? 'request_count') as KeyStatisticsSort)
                }
                items={sorts}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {sorts.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>{t('Sort order')}</FieldLabel>
              <Select
                value={sortOrder}
                onValueChange={(value) =>
                  setSortOrder((value ?? 'desc') as 'asc' | 'desc')
                }
                items={[
                  { value: 'desc', label: t('Descending') },
                  { value: 'asc', label: t('Ascending') },
                ]}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value='desc'>{t('Descending')}</SelectItem>
                    <SelectItem value='asc'>{t('Ascending')}</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>{t('Minimum tokens per request')}</FieldLabel>
              <Input
                type='number'
                min={0}
                value={minTokens}
                onChange={(event) => setMinTokens(event.target.value)}
              />
            </Field>
          </div>
          <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
            <Field>
              <FieldLabel>{t('Model contains')}</FieldLabel>
              <Input
                value={model}
                onChange={(event) => setModel(event.target.value)}
                placeholder={t('Optional model keyword')}
              />
            </Field>
            <Field>
              <FieldLabel>{t('User ID')}</FieldLabel>
              <Input
                type='number'
                min={1}
                value={userId}
                onChange={(event) => setUserId(event.target.value)}
                disabled={!allUsers}
                placeholder={t('Optional')}
              />
            </Field>
            <Field>
              <FieldLabel>{t('Exclude user ID')}</FieldLabel>
              <Input
                type='number'
                min={1}
                value={excludeUserId}
                onChange={(event) => setExcludeUserId(event.target.value)}
                disabled={!allUsers}
                placeholder={t('Optional')}
              />
            </Field>
            <div className='flex flex-col justify-end gap-2 sm:flex-row'>
              <Button
                variant='outline'
                onClick={() => {
                  if (validate()) exportMutation.mutate(query)
                }}
                disabled={exportMutation.isPending}
              >
                <Download />
                {t('Export CSV')}
              </Button>
              <Button
                onClick={() => {
                  if (validate()) queryMutation.mutate(query)
                }}
                disabled={queryMutation.isPending}
              >
                <BarChart3 />
                {queryMutation.isPending
                  ? t('Loading...')
                  : t('Query statistics')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
      {result ? (
        <>
          <StatisticsTotals data={result} />
          <Card
            data-card-hover='false'
            className='min-w-0 max-w-full overflow-hidden'
          >
            <CardHeader className='border-b'>
              <CardTitle className='text-base'>
                {t('Statistics results')}
              </CardTitle>
              <CardDescription>
                {t('Full filtered quota total: {{quota}}', {
                  quota: formatQuota(result.totals.quota),
                })}
              </CardDescription>
            </CardHeader>
            <CardContent className='min-w-0 max-w-full p-0'>
              <div className='min-w-0 max-w-full overflow-x-auto'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('Name')}</TableHead>
                      <TableHead className='text-right'>
                        {t('Requests')}
                      </TableHead>
                      <TableHead className='text-right'>
                        {t('Input Tokens')}
                      </TableHead>
                      <TableHead className='text-right'>
                        {t('Output Tokens')}
                      </TableHead>
                      <TableHead className='text-right'>{t('Quota')}</TableHead>
                      <TableHead className='text-right'>
                        {t('Unique users')}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {result.items.length > 0 ? (
                      result.items.map((row) => (
                        <TableRow key={`${row.name}-${row.request_count}`}>
                          <TableCell className='font-medium'>
                            {row.name}
                          </TableCell>
                          <TableCell className='text-right tabular-nums'>
                            {numberFormatter.format(row.request_count)}
                          </TableCell>
                          <TableCell className='text-right tabular-nums'>
                            {numberFormatter.format(row.prompt_tokens)}
                          </TableCell>
                          <TableCell className='text-right tabular-nums'>
                            {numberFormatter.format(row.completion_tokens)}
                          </TableCell>
                          <TableCell className='text-right tabular-nums'>
                            {formatQuota(row.quota)}
                          </TableCell>
                          <TableCell className='text-right tabular-nums'>
                            {numberFormatter.format(row.unique_users)}
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell
                          colSpan={6}
                          className='text-muted-foreground h-28 text-center'
                        >
                          {t('No matching statistics found.')}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </>
      ) : (
        <Alert>
          <Clock3 />
          <AlertDescription>
            {t('Choose filters and query to view aggregated usage statistics.')}
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}

export function KeyBatchOperations() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const isAdmin = Boolean(user?.role && user.role >= ROLE.ADMIN)
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Key Batch Operations')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='h-full min-h-0 min-w-0 max-w-full overflow-x-hidden overflow-y-auto pr-1'>
          <Tabs
            defaultValue='operations'
            className='grid min-h-full min-w-0 max-w-full content-start gap-4'
          >
            <TabsList className='w-fit'>
              <TabsTrigger value='operations'>
                {t('Batch operations')}
              </TabsTrigger>
              <TabsTrigger value='statistics'>
                {t('Log statistics')}
              </TabsTrigger>
            </TabsList>
            <TabsContent value='operations' className='min-w-0 max-w-full'>
              <BatchOperationsTab isAdmin={isAdmin} />
            </TabsContent>
            <TabsContent value='statistics' className='min-w-0 max-w-full'>
              <StatisticsTab isAdmin={isAdmin} />
            </TabsContent>
          </Tabs>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
