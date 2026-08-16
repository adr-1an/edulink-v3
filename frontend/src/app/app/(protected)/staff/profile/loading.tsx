import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-4">
            <Skeleton className="h-9 w-20" />
            <Skeleton className="mx-auto h-10 w-36" />
            <div className="mx-auto space-y-4 md:w-1/3">
                <Skeleton className="h-32 rounded-2xl" />
                <Skeleton className="h-24 rounded-2xl" />
                <Skeleton className="h-96 rounded-2xl" />
                <Skeleton className="h-64 rounded-2xl" />
            </div>
        </div>
    )
}
