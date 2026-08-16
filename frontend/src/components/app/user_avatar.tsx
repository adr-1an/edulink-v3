"use client"

import {useCallback, useState, useSyncExternalStore} from "react"
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar"
import {
    cacheProfilePictureURL,
    getCachedProfilePictureURL,
    invalidateProfilePictureCache,
} from "@/lib/profile_picture_cache"
import {cn} from "@/lib/utils"

function initials(name: string) {
    return name.trim().split(/\s+/).slice(0, 2)
        .map((part) => part[0]?.toUpperCase())
        .join("") || "?"
}

const subscribe = () => () => {}

export default function UserAvatar({
    name,
    src,
    cacheKey,
    className,
    fallbackClassName,
}: {
    name: string
    src?: string | null
    cacheKey?: string
    className?: string
    fallbackClassName?: string
}) {
    const [failedURL, setFailedURL] = useState<string | null>(null)
    const getSnapshot = useCallback(() => {
        if (!src || failedURL === src) return null
        return cacheKey ? getCachedProfilePictureURL(cacheKey, src) : src
    }, [cacheKey, failedURL, src])
    const resolvedSrc = useSyncExternalStore(subscribe, getSnapshot, () => null)

    return (
        <Avatar className={className}>
            {resolvedSrc && (
                <AvatarImage
                    src={resolvedSrc}
                    alt={`${name}'s profile picture`}
                    onLoadingStatusChange={(status) => {
                        if (status === "loaded" && cacheKey) {
                            cacheProfilePictureURL(cacheKey, resolvedSrc)
                        }
                        if (status === "error") {
                            if (cacheKey) invalidateProfilePictureCache(cacheKey)
                            setFailedURL(resolvedSrc)
                        }
                    }}
                />
            )}
            <AvatarFallback className={cn("text-xs", fallbackClassName)}>
                {initials(name)}
            </AvatarFallback>
        </Avatar>
    )
}
