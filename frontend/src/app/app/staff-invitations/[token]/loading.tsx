import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="w-full max-w-xl space-y-5 rounded-2xl border bg-card p-6 shadow-xl/10 sm:p-8">
                <Skeleton className="size-11 rounded-2xl" />
                <Skeleton className="h-9 w-64 max-w-full" />
                <Skeleton className="h-16 rounded-xl" />
                <div className="space-y-3">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-4 w-4/5" />
                </div>
                <div className="flex justify-end gap-2">
                    <Skeleton className="h-9 w-28" />
                    <Skeleton className="h-9 w-32" />
                </div>
        </div>
    )
}
