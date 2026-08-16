import {Card, CardContent, CardHeader} from "@/components/ui/card"
import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="mx-auto w-full max-w-5xl space-y-5">
            <Skeleton className="h-8 w-20" />
            <Card className="overflow-hidden">
                <CardHeader className="gap-6 px-5 pb-5 pt-10 sm:px-8 lg:px-12">
                    <Skeleton className="h-12 w-4/5" />
                    <Skeleton className="h-4 w-64" />
                </CardHeader>
                <CardContent className="space-y-3 border-t px-5 py-10 sm:px-8 lg:px-12">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-4 w-11/12" />
                    <Skeleton className="h-4 w-4/5" />
                    <Skeleton className="h-4 w-10/12" />
                </CardContent>
            </Card>
        </div>
    )
}
