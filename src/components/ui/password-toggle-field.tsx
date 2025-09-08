"use client"

import * as React from "react"
import * as PasswordToggleFieldPrimitive from "@radix-ui/react-password-toggle-field"
import { Eye, EyeOff, Copy, Check } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

const PasswordToggleField = ({ children, ...props }: React.ComponentPropsWithoutRef<typeof PasswordToggleFieldPrimitive.Root>) => (
  <div className="relative">
    <PasswordToggleFieldPrimitive.Root
      {...props}
    >
      {children}
    </PasswordToggleFieldPrimitive.Root>
  </div>
)

const PasswordToggleFieldInput = ({ className, ...props }: React.ComponentPropsWithoutRef<typeof PasswordToggleFieldPrimitive.Input>) => (
  <PasswordToggleFieldPrimitive.Input
    className={cn(
      "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 pr-20",
      className
    )}
    {...props}
  />
)

const PasswordToggleFieldToggle = ({ className, ...props }: React.ComponentPropsWithoutRef<typeof PasswordToggleFieldPrimitive.Toggle>) => (
  <PasswordToggleFieldPrimitive.Toggle
    className={cn(
      "absolute right-2 top-1/2 -translate-y-1/2 h-6 w-6 p-0 hover:bg-accent hover:text-accent-foreground rounded-sm flex items-center justify-center",
      className
    )}
    {...props}
  >
    <PasswordToggleFieldPrimitive.Icon
      visible={<Eye className="h-4 w-4" />}
      hidden={<EyeOff className="h-4 w-4" />}
    />
  </PasswordToggleFieldPrimitive.Toggle>
)

interface PasswordToggleFieldWithCopyProps {
  value: string
  placeholder?: string
  className?: string
  readOnly?: boolean
}

const PasswordToggleFieldWithCopy = React.forwardRef<
  HTMLDivElement,
  PasswordToggleFieldWithCopyProps
>(({ value, placeholder = "Password", className, readOnly = true, ...props }, ref) => {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy password:', error)
    }
  }

  return (
    <div ref={ref} className={cn("relative", className)} {...props}>
      <PasswordToggleField>
        <PasswordToggleFieldInput
          value={value}
          placeholder={placeholder}
          readOnly={readOnly}
        />
        <PasswordToggleFieldToggle />
      </PasswordToggleField>
      
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={handleCopy}
        className="absolute right-10 top-1/2 -translate-y-1/2 h-6 w-6 p-0 hover:bg-accent hover:text-accent-foreground rounded-sm"
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
PasswordToggleFieldWithCopy.displayName = "PasswordToggleFieldWithCopy"

export {
  PasswordToggleField,
  PasswordToggleFieldInput,
  PasswordToggleFieldToggle,
  PasswordToggleFieldWithCopy,
}
