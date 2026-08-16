import {PageHeadingSkeleton, SkeletonCardGrid} from "@/components/app/loading_skeletons"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <PageHeadingSkeleton />
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {Array.from({length: 3}, (_, index) => (
                    <Skeleton className="h-28 rounded-2xl" key={index} />
                ))}
            </div>
            <div className="space-y-3">
                <PageHeadingSkeleton action />
                <SkeletonCardGrid />
            </div>
        </div>
    )
}
