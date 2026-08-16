import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <div className="space-y-2">
                <Skeleton className="h-10 w-48" />
                <Skeleton className="h-5 w-96 max-w-full" />
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
                <Skeleton className="h-9 w-full max-w-md" />
                <Skeleton className="h-9 w-full sm:w-40" />
            </div>
            <Skeleton className="h-80 w-full rounded-xl" />
        </div>
    )
}
