import {PageHeadingSkeleton, SkeletonRows} from "@/components/app/loading_skeletons"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <PageHeadingSkeleton action />
            <div className="overflow-hidden rounded-xl border">
                <div className="flex items-center gap-3 border-b bg-muted/35 p-4">
                    <Skeleton className="size-9 rounded-lg" />
                    <div className="space-y-2">
                        <Skeleton className="h-4 w-32" />
                        <Skeleton className="h-3 w-48" />
                    </div>
                </div>
                <div className="p-3">
                    <SkeletonRows count={5} className="h-16" />
                </div>
            </div>
        </div>
    )
}
