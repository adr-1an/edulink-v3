import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <Skeleton className="h-32 rounded-3xl" />
            <Skeleton className="h-20 rounded-2xl" />
            <div className="grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(19rem,0.8fr)]">
                <Skeleton className="h-72 rounded-2xl" />
                <Skeleton className="h-72 rounded-2xl" />
            </div>
        </div>
    )
}
