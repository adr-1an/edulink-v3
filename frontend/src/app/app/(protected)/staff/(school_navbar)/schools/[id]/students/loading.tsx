import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <div className="flex items-end justify-between gap-4">
                <div className="space-y-2"><Skeleton className="h-9 w-40" /><Skeleton className="h-5 w-72" /></div>
                <Skeleton className="h-9 w-32" />
            </div>
            <Skeleton className="h-10 max-w-md" />
            <div className="space-y-3">
                {Array.from({length: 5}).map((_, index) => <Skeleton className="h-20 w-full rounded-xl" key={index} />)}
            </div>
        </div>
    )
}
