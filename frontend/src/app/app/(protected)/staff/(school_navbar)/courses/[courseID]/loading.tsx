import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-8">
            <div className="space-y-3">
                <Skeleton className="h-8 w-36" />
                <Skeleton className="h-10 w-72 max-w-full" />
                <Skeleton className="h-5 w-96 max-w-full" />
            </div>
            <div className="max-w-4xl space-y-4">
                <div className="flex items-center justify-between gap-4">
                    <Skeleton className="h-8 w-28" />
                    <Skeleton className="h-9 w-28" />
                </div>
                <Skeleton className="h-48 w-full rounded-2xl" />
                <Skeleton className="h-48 w-full rounded-2xl" />
            </div>
        </div>
    )
}
