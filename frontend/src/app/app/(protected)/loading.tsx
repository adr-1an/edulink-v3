import {PageHeadingSkeleton, SkeletonCardGrid} from "@/components/app/loading_skeletons"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6 pt-12 md:pt-0">
            <div className="mx-auto max-w-xl space-y-3 text-center">
                <Skeleton className="mx-auto h-10 w-72 max-w-full" />
                <Skeleton className="mx-auto h-10 w-full" />
                <Skeleton className="mx-auto h-3 w-28" />
            </div>
            <div className="space-y-4">
                <PageHeadingSkeleton />
                <SkeletonCardGrid cardClassName="h-36" />
            </div>
        </div>
    )
}
