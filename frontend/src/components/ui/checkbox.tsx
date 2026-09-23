"use client";

import { Checkbox as CheckboxPrimitive } from "@base-ui/react/checkbox";
import type React from "react";
import { cn } from "@/lib/utils";

export function Checkbox({
  className,
  ...props
}: CheckboxPrimitive.Root.Props): React.ReactElement {
  return (
    <CheckboxPrimitive.Root
      className={cn(
        "relative inline-flex size-5 shrink-0 items-center justify-center rounded-md border border-input bg-background outline-none transition-[background-color,border-color,box-shadow] focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/25 aria-invalid:border-destructive/50 data-checked:border-primary data-checked:bg-primary data-disabled:cursor-not-allowed data-disabled:opacity-50 sm:size-4.5",
        className,
      )}
      data-slot="checkbox"
      {...props}
    >
      <CheckboxPrimitive.Indicator
        className="absolute inset-0 flex items-center justify-center rounded-[inherit] text-primary-foreground data-unchecked:hidden data-indeterminate:text-primary-foreground"
        data-slot="checkbox-indicator"
        render={(
          props: React.ComponentProps<"span">,
          state: CheckboxPrimitive.Indicator.State,
        ) => (
          <span {...props}>
            {state.indeterminate ? (
              <svg
                aria-hidden="true"
                className="size-3.5 sm:size-3"
                fill="none"
                height="24"
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="3"
                viewBox="0 0 24 24"
                width="24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M5.252 12h13.496" />
              </svg>
            ) : (
              <svg
                aria-hidden="true"
                className="size-3.5 sm:size-3"
                fill="none"
                height="24"
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="3"
                viewBox="0 0 24 24"
                width="24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M5.252 12.7 10.2 18.63 18.748 5.37" />
              </svg>
            )}
          </span>
        )}
      />
    </CheckboxPrimitive.Root>
  );
}

export { CheckboxPrimitive };
