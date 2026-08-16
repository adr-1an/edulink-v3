import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="w-full max-w-lg overflow-hidden rounded-2xl border bg-card shadow-xl/10">
            <div className="space-y-4 border-b bg-muted/20 px-5 py-6 sm:px-7">
                <Skeleton className="size-11 rounded-2xl" />
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-9 w-64 max-w-full" />
                <Skeleton className="h-4 w-full" />
            </div>
            <div className="space-y-5 px-5 py-6 sm:px-7">
                <div className="space-y-2">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-10 w-full" />
                </div>
                <div className="space-y-2">
                    <Skeleton className="h-4 w-20" />
                    <Skeleton className="h-10 w-full" />
                </div>
                <Skeleton className="h-10 w-full" />
            </div>
        </div>
    )
}
