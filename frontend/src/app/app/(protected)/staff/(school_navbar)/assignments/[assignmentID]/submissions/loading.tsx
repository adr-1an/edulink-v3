import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <Skeleton className="h-9 w-36" />
            <div className="space-y-3">
                <div className="flex gap-2">
                    <Skeleton className="h-6 w-28 rounded-full" />
                    <Skeleton className="h-6 w-20 rounded-full" />
                </div>
                <Skeleton className="h-10 w-80 max-w-full" />
                <Skeleton className="h-5 w-[32rem] max-w-full" />
            </div>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <Skeleton className="h-10 w-full sm:max-w-md" />
                <Skeleton className="h-5 w-40" />
            </div>
            <Skeleton className="h-96 w-full rounded-2xl" />
        </div>
    )
}
