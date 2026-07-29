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
import { Check, Copy, KeyRound, Search } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

import { copyText } from '../utils'
import { ActivationDialog } from './activation-dialog'
import { KeyQueryDialog } from './key-query-dialog'

export function XuancatHomeAccessPanel() {
  const { t } = useTranslation()
  const [activationOpen, setActivationOpen] = useState(false)
  const [keyQueryOpen, setKeyQueryOpen] = useState(false)
  const [copied, setCopied] = useState(false)
  const apiUrl = `${window.location.origin}/v1`

  const copyApiUrl = async () => {
    await copyText(apiUrl)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <div
        className='landing-animate-fade-up mx-auto mt-12 max-w-6xl opacity-0'
        style={{ animationDelay: '380ms' }}
      >
        <div className='grid gap-3 sm:grid-cols-2'>
          <Button
            type='button'
            variant='outline'
            className='border-border/50 bg-background/55 hover:border-border hover:bg-muted/40 h-24 flex-col gap-2 rounded-2xl text-sm font-medium shadow-xs backdrop-blur-sm'
            onClick={() => setActivationOpen(true)}
          >
            <KeyRound className='size-6 text-blue-500' aria-hidden='true' />
            {t('Activate or renew a key')}
          </Button>
          <Button
            type='button'
            variant='outline'
            className='border-border/50 bg-background/55 hover:border-border hover:bg-muted/40 h-24 flex-col gap-2 rounded-2xl text-sm font-medium shadow-xs backdrop-blur-sm'
            onClick={() => setKeyQueryOpen(true)}
          >
            <Search className='size-6 text-blue-500' aria-hidden='true' />
            {t('Query an API key')}
          </Button>
        </div>

        <div className='border-border/50 bg-background/55 mt-3 rounded-2xl border p-4 shadow-xs backdrop-blur-sm sm:p-5'>
          <div className='mb-4'>
            <h2 className='text-base font-semibold'>{t('API base URL')}</h2>
            <p className='text-muted-foreground/80 mt-1 text-sm'>
              {t('Use this address in OpenAI-compatible clients.')}
            </p>
          </div>
          <div className='flex items-center gap-2'>
            <code className='bg-muted/60 min-w-0 flex-1 overflow-x-auto rounded-xl px-3.5 py-2.5 text-sm'>
              {apiUrl}
            </code>
            <Button
              type='button'
              size='icon'
              variant='outline'
              className='border-border/50 bg-background/70 shrink-0 rounded-xl'
              aria-label={copied ? t('Copied') : t('Copy API base URL')}
              onClick={copyApiUrl}
            >
              {copied ? (
                <Check aria-hidden='true' />
              ) : (
                <Copy aria-hidden='true' />
              )}
            </Button>
          </div>
        </div>
      </div>

      <ActivationDialog
        open={activationOpen}
        onOpenChange={setActivationOpen}
      />
      <KeyQueryDialog open={keyQueryOpen} onOpenChange={setKeyQueryOpen} />
    </>
  )
}
