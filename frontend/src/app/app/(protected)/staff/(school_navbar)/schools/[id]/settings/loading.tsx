import {PageHeadingSkeleton} from "@/components/app/loading_skeletons"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <PageHeadingSkeleton action />
            {Array.from({length: 3}, (_, index) => (
                <section className="space-y-3" key={index}>
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className={index === 1 ? "h-72 rounded-2xl" : "h-56 rounded-2xl"} />
                </section>
            ))}
        </div>
    )
}
