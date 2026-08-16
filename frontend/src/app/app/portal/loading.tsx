import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="mx-auto w-full max-w-7xl space-y-6 px-4 py-8 sm:px-6 lg:px-8">
            <Skeleton className="h-36 rounded-3xl" />
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                {Array.from({length: 6}, (_, index) => <Skeleton className="h-44 rounded-2xl" key={index} />)}
            </div>
        </div>
    )
}
