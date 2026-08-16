import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <Skeleton className="h-64 rounded-3xl" />
            <Skeleton className="h-9 w-64 rounded-lg" />
            <div className="grid gap-6 xl:grid-cols-[minmax(0,2fr)_minmax(17rem,0.8fr)]">
                <div className="space-y-3">
                    <Skeleton className="h-12 w-48" />
                    <Skeleton className="h-52 rounded-2xl" />
                    <Skeleton className="h-44 rounded-2xl" />
                </div>
                <Skeleton className="h-64 rounded-2xl" />
            </div>
        </div>
    )
}
