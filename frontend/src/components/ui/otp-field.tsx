"use client";

import { OTPField as OTPFieldPrimitive } from "@base-ui/react/otp-field";
import type * as React from "react";
import { cn } from "@/lib/utils";
import { Separator } from "@/components/ui/separator";

export function OTPField({
  className,
  size = "default",
  ...props
}: React.ComponentProps<typeof OTPFieldPrimitive.Root> & {
  size?: "default" | "lg";
}): React.ReactElement {
  return (
    <OTPFieldPrimitive.Root
      className={cn(
        "flex items-center gap-2 has-disabled:opacity-50",
        className,
      )}
      data-size={size}
      data-slot="otp-field"
      {...props}
    />
  );
}

export function OTPFieldInput({
  className,
  ...props
}: React.ComponentProps<typeof OTPFieldPrimitive.Input>): React.ReactElement {
  return (
    <OTPFieldPrimitive.Input
      className={cn(
        "relative in-[[data-slot=otp-field][data-size=lg]]:size-12 size-11 min-w-0 rounded-xl border border-input bg-background text-center in-[[data-slot=otp-field][data-size=lg]]:text-lg text-base text-foreground in-[[data-slot=otp-field][data-size=lg]]:leading-12 leading-11 shadow-none outline-none ring-ring/20 transition-[border-color,box-shadow,background-color] focus-visible:z-10 focus-visible:border-ring focus-visible:ring-[3px] aria-invalid:border-destructive/50 aria-invalid:focus-visible:border-destructive aria-invalid:focus-visible:ring-destructive/15 sm:in-[[data-slot=otp-field][data-size=lg]]:size-10 sm:size-9 sm:in-[[data-slot=otp-field][data-size=lg]]:text-base sm:text-sm sm:in-[[data-slot=otp-field][data-size=lg]]:leading-10 sm:leading-9",
        className,
      )}
      data-slot="otp-field-input"
      spellCheck={false}
      {...props}
    />
  );
}

export function OTPFieldSeparator({
  className,
  ...props
}: React.ComponentProps<typeof Separator>): React.ReactElement {
  return (
    <OTPFieldPrimitive.Separator
      render={
        <Separator
          className={cn(
            "rounded-full bg-input data-[orientation=horizontal]:h-0.5 data-[orientation=horizontal]:w-3",
            className,
          )}
          orientation="horizontal"
          {...props}
        />
      }
    />
  );
}

export { OTPFieldPrimitive };
