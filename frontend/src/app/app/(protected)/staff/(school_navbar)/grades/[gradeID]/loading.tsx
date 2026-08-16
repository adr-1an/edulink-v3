import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <div className="flex items-end justify-between gap-4">
                <div className="space-y-2">
                    <Skeleton className="h-9 w-44" />
                    <Skeleton className="h-4 w-72" />
                </div>
                <Skeleton className="h-8 w-32" />
            </div>
            <Skeleton className="h-9 w-full max-w-md" />
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {Array.from({length: 6}, (_, index) => <Skeleton key={index} className="h-44 rounded-xl" />)}
            </div>
        </div>
    )
}
