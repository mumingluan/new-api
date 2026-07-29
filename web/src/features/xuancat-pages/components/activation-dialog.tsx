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
import { toast } from 'sonner'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

import {
  precheckActivation,
  queryActivation,
  redeemActivation,
  renewActivation,
} from '../api'
import type { ActivationQueryResponse, PrecheckResponse } from '../types'
import { copyText, formatTimestamp } from '../utils'

type ActivationDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Confirmation =
  | {
      type: 'redeem'
      precheck: PrecheckResponse
      activationCode: string
      qq: string
    }
  | {
      type: 'renew'
      precheck: PrecheckResponse
      activationCode: string
      apiKey: string
    }

function errorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback
}

export function ActivationDialog(props: ActivationDialogProps) {
  const { i18n, t } = useTranslation()
  const [activationCode, setActivationCode] = useState('')
  const [qq, setQq] = useState('')
  const [renewCode, setRenewCode] = useState('')
  const [renewKey, setRenewKey] = useState('')
  const [queryCode, setQueryCode] = useState('')
  const [createdKey, setCreatedKey] = useState('')
  const [queryResult, setQueryResult] =
    useState<ActivationQueryResponse | null>(null)
  const [confirmation, setConfirmation] = useState<Confirmation | null>(null)
  const [pending, setPending] = useState(false)

  const prepareRedeem = async () => {
    if (!activationCode.trim() || !qq.trim()) {
      toast.error(t('Activation code and QQ number are required.'))
      return
    }

    setPending(true)
    try {
      const precheck = await precheckActivation('', {
        activation_code: activationCode.trim(),
        qq: qq.trim(),
      })
      if (!precheck.valid) {
        throw new Error(precheck.error || t('Verification failed.'))
      }
      setConfirmation({
        type: 'redeem',
        precheck,
        activationCode: activationCode.trim(),
        qq: qq.trim(),
      })
    } catch (error) {
      toast.error(errorMessage(error, t('Unable to reach the server.')))
    } finally {
      setPending(false)
    }
  }

  const prepareRenew = async () => {
    if (!renewCode.trim() || !renewKey.trim()) {
      toast.error(t('API key and activation code are required.'))
      return
    }

    setPending(true)
    try {
      const precheck = await precheckActivation('', {
        activation_code: renewCode.trim(),
        api_key: renewKey.trim(),
      })
      if (!precheck.valid) {
        throw new Error(precheck.error || t('Verification failed.'))
      }
      setConfirmation({
        type: 'renew',
        precheck,
        activationCode: renewCode.trim(),
        apiKey: renewKey.trim(),
      })
    } catch (error) {
      toast.error(errorMessage(error, t('Unable to reach the server.')))
    } finally {
      setPending(false)
    }
  }

  const executeConfirmation = async () => {
    if (!confirmation) return
    setPending(true)

    try {
      if (confirmation.type === 'redeem') {
        const result = await redeemActivation(
          '',
          confirmation.activationCode,
          confirmation.qq
        )
        setCreatedKey(result.api_key)
        toast.success(
          t('API key created. It expires on {{date}}.', {
            date: formatTimestamp(result.expired_time, i18n.language),
          })
        )
      } else {
        const result = await renewActivation(
          '',
          confirmation.activationCode,
          confirmation.apiKey
        )
        toast.success(
          t('API key renewed until {{date}}.', {
            date: formatTimestamp(result.new_expired_time, i18n.language),
          })
        )
      }
      setConfirmation(null)
    } catch (error) {
      toast.error(errorMessage(error, t('Unable to reach the server.')))
    } finally {
      setPending(false)
    }
  }

  const runQuery = async () => {
    if (!queryCode.trim()) {
      toast.error(t('Enter an activation code.'))
      return
    }

    setPending(true)
    setQueryResult(null)
    try {
      setQueryResult(await queryActivation('', queryCode.trim()))
    } catch (error) {
      toast.error(errorMessage(error, t('Unable to reach the server.')))
    } finally {
      setPending(false)
    }
  }

  const copyKey = async (key: string) => {
    try {
      await copyText(key)
      toast.success(t('Copied to clipboard'))
    } catch {
      toast.error(t('Copy failed'))
    }
  }

  return (
    <>
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>{t('Activation code service')}</DialogTitle>
            <DialogDescription>
              {t('Create, renew, or recover a timed API key.')}
            </DialogDescription>
          </DialogHeader>

          <Tabs defaultValue='create'>
            <TabsList className='grid w-full grid-cols-3'>
              <TabsTrigger value='create'>{t('Create key')}</TabsTrigger>
              <TabsTrigger value='renew'>{t('Renew key')}</TabsTrigger>
              <TabsTrigger value='query'>{t('Recover key')}</TabsTrigger>
            </TabsList>

            <TabsContent value='create' className='space-y-4 pt-2'>
              <div className='space-y-2'>
                <Label htmlFor='xuancat-create-code'>
                  {t('Activation code')}
                </Label>
                <Input
                  id='xuancat-create-code'
                  value={activationCode}
                  onChange={(event) => setActivationCode(event.target.value)}
                  placeholder={t('Enter your activation code')}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='xuancat-create-qq'>{t('QQ number')}</Label>
                <Input
                  id='xuancat-create-qq'
                  value={qq}
                  onChange={(event) => setQq(event.target.value)}
                  placeholder={t('Enter your QQ number')}
                />
              </div>
              <Button
                className='w-full'
                disabled={pending}
                onClick={prepareRedeem}
              >
                {pending ? t('Verifying...') : t('Create key')}
              </Button>
              {createdKey && (
                <Alert>
                  <AlertTitle>{t('Your API key')}</AlertTitle>
                  <AlertDescription className='space-y-3'>
                    <code className='block break-all'>{createdKey}</code>
                    <Button
                      size='sm'
                      variant='outline'
                      onClick={() => copyKey(createdKey)}
                    >
                      {t('Copy')}
                    </Button>
                  </AlertDescription>
                </Alert>
              )}
            </TabsContent>

            <TabsContent value='renew' className='space-y-4 pt-2'>
              <div className='space-y-2'>
                <Label htmlFor='xuancat-renew-key'>{t('API key')}</Label>
                <Input
                  id='xuancat-renew-key'
                  value={renewKey}
                  onChange={(event) => setRenewKey(event.target.value)}
                  placeholder='sk-...'
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='xuancat-renew-code'>
                  {t('Activation code')}
                </Label>
                <Input
                  id='xuancat-renew-code'
                  value={renewCode}
                  onChange={(event) => setRenewCode(event.target.value)}
                  placeholder={t('Enter your activation code')}
                />
              </div>
              <Button
                className='w-full'
                disabled={pending}
                onClick={prepareRenew}
              >
                {pending ? t('Verifying...') : t('Renew key')}
              </Button>
            </TabsContent>

            <TabsContent value='query' className='space-y-4 pt-2'>
              <div className='space-y-2'>
                <Label htmlFor='xuancat-query-code'>
                  {t('Activation code')}
                </Label>
                <Input
                  id='xuancat-query-code'
                  value={queryCode}
                  onChange={(event) => setQueryCode(event.target.value)}
                  placeholder={t('Enter your activation code')}
                />
              </div>
              <Button className='w-full' disabled={pending} onClick={runQuery}>
                {pending ? t('Querying...') : t('Recover key')}
              </Button>
              {queryResult && (
                <Alert>
                  <AlertTitle>{t('Activation record')}</AlertTitle>
                  <AlertDescription className='space-y-2'>
                    <p>
                      {t('Used at')}: {queryResult.used_time}
                    </p>
                    <p>
                      {t('Operation')}:{' '}
                      {queryResult.action === 'create'
                        ? t('Create')
                        : t('Renew key')}
                    </p>
                    <code className='block break-all'>
                      {queryResult.api_key}
                    </code>
                    <Button
                      size='sm'
                      variant='outline'
                      onClick={() => copyKey(queryResult.api_key)}
                    >
                      {t('Copy')}
                    </Button>
                  </AlertDescription>
                </Alert>
              )}
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>

      <Dialog
        open={confirmation !== null}
        onOpenChange={(open) => {
          if (!open) setConfirmation(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Confirm activation')}</DialogTitle>
            <DialogDescription>
              {t('Check the information before consuming the activation code.')}
            </DialogDescription>
          </DialogHeader>
          {confirmation && (
            <dl className='grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm'>
              <dt className='text-muted-foreground'>{t('IP address')}</dt>
              <dd>{confirmation.precheck.client_ip}</dd>
              <dt className='text-muted-foreground'>{t('Channel')}</dt>
              <dd>{confirmation.precheck.channel}</dd>
              <dt className='text-muted-foreground'>{t('Duration')}</dt>
              <dd>
                {t('{{count}} days', {
                  count: confirmation.precheck.days,
                })}
              </dd>
              <dt className='text-muted-foreground'>{t('New expiry')}</dt>
              <dd>
                {formatTimestamp(
                  confirmation.precheck.new_expired_time,
                  i18n.language
                )}
              </dd>
            </dl>
          )}
          <DialogFooter>
            <Button variant='outline' onClick={() => setConfirmation(null)}>
              {t('Cancel')}
            </Button>
            <Button disabled={pending} onClick={executeConfirmation}>
              {pending ? t('Processing...') : t('Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
