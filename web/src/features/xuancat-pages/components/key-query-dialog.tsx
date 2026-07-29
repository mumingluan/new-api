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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { getSubscription, getUsage, getUsageLogs } from '../api'
import type { TokenSummary, UsageLog } from '../types'
import { formatQuota, formatTimestamp, getUsageDateRange } from '../utils'

type KeyQueryDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function KeyQueryDialog(props: KeyQueryDialogProps) {
  const { i18n, t } = useTranslation()
  const [apiKey, setApiKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [summary, setSummary] = useState<TokenSummary | null>(null)
  const [logs, setLogs] = useState<UsageLog[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const queryKey = async () => {
    const key = apiKey.trim()
    setError('')
    setSummary(null)
    setLogs([])

    if (!key) {
      setError(t('Enter an API key.'))
      return
    }
    if (!/^sk-[a-zA-Z0-9]{48}$/.test(key)) {
      setError(t('The API key format is invalid.'))
      return
    }

    setLoading(true)
    try {
      const { startDate, endDate } = getUsageDateRange()
      const [subscription, usageData, logData] = await Promise.all([
        getSubscription('', key),
        getUsage('', key, startDate, endDate),
        getUsageLogs('', key),
      ])
      const balance = subscription.hard_limit_usd
      const usage = usageData.total_usage / 100
      const unlimited = balance === 100_000_000
      const usageLogs = (logData.success ? (logData.data ?? []) : []).filter(
        (log) => log.type === 0 || log.type === 2
      )
      const tokenName =
        subscription.token_name ||
        usageLogs.find((log) => log.token_name)?.token_name ||
        t('Unknown')

      setSummary({
        name: tokenName,
        balance: unlimited ? t('Unlimited') : balance.toFixed(3),
        remaining: unlimited ? t('No limit') : (balance - usage).toFixed(3),
        used: unlimited ? t('Not calculated') : usage.toFixed(3),
        expiry:
          subscription.access_until && subscription.access_until > 0
            ? formatTimestamp(subscription.access_until, i18n.language)
            : t('Never expires'),
      })
      setLogs(usageLogs)
      setPage(1)
    } catch (queryError) {
      setError(
        queryError instanceof Error && queryError.message
          ? queryError.message
          : t('Unable to query this API key.')
      )
    } finally {
      setLoading(false)
    }
  }

  const pageCount = Math.max(1, Math.ceil(logs.length / pageSize))
  const safePage = Math.min(page, pageCount)
  const startIndex = (safePage - 1) * pageSize
  const visibleLogs = logs.slice(startIndex, startIndex + pageSize)

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-5xl'>
        <DialogHeader>
          <DialogTitle>{t('API key query')}</DialogTitle>
          <DialogDescription>
            {t('View the balance, expiry, and recent calls for an API key.')}
          </DialogDescription>
        </DialogHeader>

        <div className='flex flex-col gap-3 sm:flex-row sm:items-end'>
          <div className='flex-1 space-y-2'>
            <Label htmlFor='xuancat-query-key'>{t('API key')}</Label>
            <Input
              id='xuancat-query-key'
              value={apiKey}
              onChange={(event) => setApiKey(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') queryKey()
              }}
              placeholder='sk-...'
            />
          </div>
          <Button disabled={loading} onClick={queryKey}>
            {loading ? t('Querying...') : t('Query')}
          </Button>
        </div>

        {error && (
          <Alert variant='destructive'>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {summary && (
          <div className='space-y-3'>
            <Card size='sm'>
              <CardHeader className='gap-1 sm:flex-row sm:items-center sm:justify-between'>
                <CardDescription>{t('Token Name')}</CardDescription>
                <CardTitle className='text-base break-all'>
                  {summary.name}
                </CardTitle>
              </CardHeader>
            </Card>
            <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
              {[
                [t('Total quota'), summary.balance],
                [t('Remaining quota'), summary.remaining],
                [t('Used quota'), summary.used],
                [t('Expiry'), summary.expiry],
              ].map(([label, value]) => (
                <Card key={label} size='sm'>
                  <CardHeader>
                    <CardDescription>{label}</CardDescription>
                    <CardTitle>{value}</CardTitle>
                  </CardHeader>
                </Card>
              ))}
            </div>
          </div>
        )}

        {summary && (
          <Card>
            <CardHeader>
              <CardTitle>{t('Recent calls')}</CardTitle>
              <CardDescription>
                {t('Quota conversion: $1 = 500,000 tokens')}
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              {logs.length === 0 ? (
                <p className='text-muted-foreground py-8 text-center'>
                  {t('No call records')}
                </p>
              ) : (
                <>
                  <div className='overflow-x-auto'>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('Time')}</TableHead>
                          <TableHead>{t('Model')}</TableHead>
                          <TableHead>{t('Group')}</TableHead>
                          <TableHead>{t('Duration')}</TableHead>
                          <TableHead>{t('Prompt')}</TableHead>
                          <TableHead>{t('Completion')}</TableHead>
                          <TableHead>{t('Cost')}</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {visibleLogs.map((log, index) => (
                          <TableRow
                            key={
                              log.id ??
                              `${log.created_at}-${log.model_name}-${index}`
                            }
                          >
                            <TableCell className='whitespace-nowrap'>
                              {formatTimestamp(log.created_at, i18n.language)}
                            </TableCell>
                            <TableCell>
                              <Badge variant='secondary'>
                                {log.model_name}
                              </Badge>
                            </TableCell>
                            <TableCell>{log.group || 'default'}</TableCell>
                            <TableCell>
                              {t('{{count}} seconds', {
                                count: Number(log.use_time),
                              })}{' '}
                              <Badge variant='outline'>
                                {log.is_stream
                                  ? t('Streaming')
                                  : t('Non-streaming')}
                              </Badge>
                            </TableCell>
                            <TableCell>{log.prompt_tokens || '-'}</TableCell>
                            <TableCell>
                              {log.completion_tokens || '-'}
                            </TableCell>
                            <TableCell>${formatQuota(log.quota)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>

                  <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
                    <p className='text-muted-foreground text-sm'>
                      {t('Showing {{start}}–{{end}} of {{total}}', {
                        start: startIndex + 1,
                        end: Math.min(startIndex + pageSize, logs.length),
                        total: logs.length,
                      })}
                    </p>
                    <div className='flex items-center gap-2'>
                      <NativeSelect
                        aria-label={t('Rows per page')}
                        value={pageSize}
                        onChange={(event) => {
                          setPageSize(Number(event.target.value))
                          setPage(1)
                        }}
                      >
                        {[10, 20, 50, 100].map((size) => (
                          <NativeSelectOption key={size} value={size}>
                            {size}
                          </NativeSelectOption>
                        ))}
                      </NativeSelect>
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={safePage === 1}
                        onClick={() => setPage((current) => current - 1)}
                      >
                        {t('Previous')}
                      </Button>
                      <span className='text-sm'>
                        {safePage} / {pageCount}
                      </span>
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={safePage === pageCount}
                        onClick={() => setPage((current) => current + 1)}
                      >
                        {t('Next')}
                      </Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        )}
      </DialogContent>
    </Dialog>
  )
}
