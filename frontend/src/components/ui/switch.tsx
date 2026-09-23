"use client";

import { Switch as SwitchPrimitive } from "@base-ui/react/switch";
import type React from "react";
import { cn } from "@/lib/utils";

export function Switch({
  className,
  ...props
}: SwitchPrimitive.Root.Props): React.ReactElement {
  return (
    <SwitchPrimitive.Root
      className={cn(
        "inline-flex h-[calc(var(--thumb-size)+4px)] w-[calc(var(--thumb-size)*2)] shrink-0 items-center rounded-full bg-input p-0.5 outline-none transition-[background-color,box-shadow] duration-200 [--thumb-size:--spacing(5)] focus-visible:ring-2 focus-visible:ring-ring/30 data-checked:bg-primary data-disabled:cursor-not-allowed data-disabled:opacity-50 sm:[--thumb-size:--spacing(4)]",
        className,
      )}
      data-slot="switch"
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          "pointer-events-none block aspect-square h-full rounded-full bg-white shadow-sm will-change-transform transition-transform duration-200 data-checked:translate-x-[calc(var(--thumb-size)-4px)]",
        )}
        data-slot="switch-thumb"
      />
    </SwitchPrimitive.Root>
  );
}

export { SwitchPrimitive };
