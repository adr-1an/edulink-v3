"use client";

import { Input as InputPrimitive } from "@base-ui/react/input";
import type * as React from "react";
import { cn } from "@/lib/utils";

export type InputProps = Omit<
  InputPrimitive.Props & React.RefAttributes<HTMLInputElement>,
  "size"
> & {
  size?: "sm" | "default" | "lg" | number;
  unstyled?: boolean;
  nativeInput?: boolean;
};

export function Input({
  className,
  size = "default",
  unstyled = false,
  nativeInput = false,
  style,
  ...props
}: InputProps): React.ReactElement {
  const inputClassName = cn(
    "h-11 w-full min-w-0 rounded-[inherit] bg-transparent px-3 leading-11 outline-none [transition:background-color_5000000s_ease-in-out_0s] placeholder:text-muted-foreground sm:h-9 sm:leading-9",
    size === "sm" &&
      "h-10 px-2.5 leading-10 sm:h-8 sm:leading-8",
    size === "lg" && "h-12 leading-12 sm:h-10 sm:leading-10",
    props.type === "search" &&
      "[&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none",
    props.type === "file" &&
      "text-muted-foreground file:me-3 file:bg-transparent file:font-medium file:text-foreground file:text-sm",
    props.type === "color" &&
      "cursor-pointer p-1 [&::-moz-color-swatch]:rounded-md [&::-moz-color-swatch]:border-0 [&::-webkit-color-swatch]:rounded-md [&::-webkit-color-swatch]:border-0 [&::-webkit-color-swatch-wrapper]:p-0",
  );

  return (
    <span
      className={
        cn(
          !unstyled &&
            "relative inline-flex w-full rounded-lg border border-input bg-background text-base text-foreground shadow-none ring-ring/20 transition-[border-color,box-shadow,background-color] has-focus-visible:border-ring has-focus-visible:ring-[3px] has-aria-invalid:border-destructive/50 has-focus-visible:has-aria-invalid:border-destructive has-focus-visible:has-aria-invalid:ring-destructive/15 has-disabled:bg-muted/50 has-disabled:opacity-60 sm:text-sm",
          className,
        ) || undefined
      }
      data-size={size}
      data-slot="input-control"
    >
      {nativeInput ? (
        <input
          className={inputClassName}
          data-slot="input"
          size={typeof size === "number" ? size : undefined}
          style={typeof style === "function" ? undefined : style}
          {...props}
        />
      ) : (
        <InputPrimitive
          className={inputClassName}
          data-slot="input"
          size={typeof size === "number" ? size : undefined}
          style={style}
          {...props}
        />
      )}
    </span>
  );
}

export { InputPrimitive };
