import {SkeletonCardGrid} from "@/components/app/loading_skeletons"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-7">
            <Skeleton className="h-44 rounded-3xl" />
            <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
                <div className="space-y-2">
                    <Skeleton className="h-6 w-28" />
                    <Skeleton className="h-4 w-44" />
                </div>
                <Skeleton className="h-9 w-full sm:w-80" />
            </div>
            <SkeletonCardGrid />
        </div>
    )
}
