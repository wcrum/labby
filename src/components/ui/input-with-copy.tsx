"use client"

import * as React from "react"
import { Copy, Check } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

interface InputWithCopyProps {
  value: string
  placeholder?: string
  className?: string
  readOnly?: boolean
  type?: string
}

const InputWithCopy = React.forwardRef<
  HTMLDivElement,
  InputWithCopyProps
>(({ value, placeholder, className, readOnly = true, type = "text", ...props }, ref) => {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy text:', error)
    }
  }

  return (
    <div ref={ref} className={cn("relative", className)} {...props}>
      <Input
        value={value}
        placeholder={placeholder}
        readOnly={readOnly}
        type={type}
        className="h-10 pr-12 !bg-background dark:!bg-background px-3 py-2 text-sm ring-offset-background"
      />
      
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={handleCopy}
        className="absolute right-2 top-1/2 -translate-y-1/2 h-6 w-6 p-0 hover:bg-accent hover:text-accent-foreground rounded-sm"
      >
        {copied ? (
          <Check className="h-4 w-4 text-green-600" />
        ) : (
          <Copy className="h-4 w-4" />
        )}
      </Button>
    </div>
  )
})
InputWithCopy.displayName = "InputWithCopy"

export { InputWithCopy }
