import {PageHeadingSkeleton, SkeletonRows} from "@/components/app/loading_skeletons"

export default function Loading() {
    return (
        <div className="space-y-6">
            <PageHeadingSkeleton action />
            <SkeletonRows />
        </div>
    )
}
