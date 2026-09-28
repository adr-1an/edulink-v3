"use client";

import { Toggle as TogglePrimitive } from "@base-ui/react/toggle";
import { cva, type VariantProps } from "class-variance-authority";
import type React from "react";
import { cn } from "@/lib/utils";

export const toggleVariants = cva(
  "relative inline-flex shrink-0 cursor-pointer select-none items-center justify-center gap-2 whitespace-nowrap rounded-md border font-medium text-base text-foreground outline-none transition-[color,background-color,border-color,box-shadow] pointer-coarse:after:absolute pointer-coarse:after:size-full pointer-coarse:after:min-h-11 pointer-coarse:after:min-w-11 hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50 data-pressed:bg-secondary data-pressed:text-secondary-foreground sm:text-sm [&_svg:not([class*='opacity-'])]:opacity-80 [&_svg:not([class*='size-'])]:size-4.5 sm:[&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:-mx-0.5 [&_svg]:shrink-0",
  {
    defaultVariants: {
      size: "default",
      variant: "default",
    },
    variants: {
      size: {
        default: "h-9 min-w-9 px-[calc(--spacing(2)-1px)] sm:h-8 sm:min-w-8",
        lg: "h-10 min-w-10 px-[calc(--spacing(2.5)-1px)] sm:h-9 sm:min-w-9",
        sm: "h-8 min-w-8 px-[calc(--spacing(1.5)-1px)] sm:h-7 sm:min-w-7",
      },
      variant: {
        default: "border-transparent",
        outline:
          "border-input bg-transparent data-pressed:border-transparent",
      },
    },
  },
);

export function Toggle({
  className,
  variant,
  size,
  ...props
}: TogglePrimitive.Props &
  VariantProps<typeof toggleVariants>): React.ReactElement {
  return (
    <TogglePrimitive
      className={cn(toggleVariants({ className, size, variant }))}
      data-slot="toggle"
      {...props}
    />
  );
}

export { TogglePrimitive };
