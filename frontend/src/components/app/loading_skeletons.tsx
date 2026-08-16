import {Skeleton} from "@/components/ui/skeleton"
import {cn} from "@/lib/utils"

export function PageHeadingSkeleton({action = false}: {action?: boolean}) {
    return (
        <div className="flex flex-wrap items-end justify-between gap-4">
            <div className="space-y-2">
                <Skeleton className="h-9 w-48" />
                <Skeleton className="h-5 w-72 max-w-full" />
            </div>
            {action && <Skeleton className="h-9 w-36" />}
        </div>
    )
}

export function SkeletonRows({
    count = 5,
    className,
}: {
    count?: number
    className?: string
}) {
    return (
        <div className="space-y-3">
            {Array.from({length: count}, (_, index) => (
                <Skeleton className={cn("h-20 w-full rounded-xl", className)} key={index} />
            ))}
        </div>
    )
}

export function SkeletonCardGrid({
    count = 6,
    cardClassName,
}: {
    count?: number
    cardClassName?: string
}) {
    return (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({length: count}, (_, index) => (
                <Skeleton className={cn("h-40 rounded-2xl", cardClassName)} key={index} />
            ))}
        </div>
    )
}
