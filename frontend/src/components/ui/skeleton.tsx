import type React from "react";
import { cn } from "@/lib/utils";

export function Skeleton({
  className,
  ...props
}: React.ComponentProps<"div">): React.ReactElement {
  return (
    <div
      className={cn(
        "animate-skeleton rounded-lg [--skeleton-highlight:--alpha(var(--color-white)/45%)] [background:linear-gradient(120deg,transparent_40%,var(--skeleton-highlight),transparent_60%)_var(--color-muted)_0_0/200%_100%_fixed] dark:[--skeleton-highlight:--alpha(var(--color-white)/5%)]",
        className,
      )}
      data-slot="skeleton"
      {...props}
    />
  );
}
