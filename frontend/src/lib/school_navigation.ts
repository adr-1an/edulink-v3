"use client"

import type {SchoolAccess} from "@/lib/school_access"

const ACTIVE_SCHOOL_STORAGE_KEY = "edulink:active-school-id"
const ACTIVE_SCHOOL_ACCESS_KEY = "edulink:active-school-access"
const ACTIVE_SCHOOL_EVENT = "edulink:active-school-changed"

interface StoredSchoolAccess {
    schoolID: string
    access: SchoolAccess
}

export function subscribeToSchoolNavigation(callback: () => void) {
    window.addEventListener("storage", callback)
    window.addEventListener(ACTIVE_SCHOOL_EVENT, callback)

    return () => {
        window.removeEventListener("storage", callback)
        window.removeEventListener(ACTIVE_SCHOOL_EVENT, callback)
    }
}

export function getActiveSchoolSnapshot() {
    const schoolID = sessionStorage.getItem(ACTIVE_SCHOOL_STORAGE_KEY)
    return schoolID && /^\d+$/.test(schoolID) ? schoolID : null
}

export function getActiveSchoolAccessSnapshot() {
    return sessionStorage.getItem(ACTIVE_SCHOOL_ACCESS_KEY)
}

export function getServerSchoolNavigationSnapshot() {
    return null
}

export function parseStoredSchoolAccess(value: string | null): StoredSchoolAccess | null {
    if (!value) return null

    try {
        const parsed: unknown = JSON.parse(value)
        if (!parsed || typeof parsed !== "object") return null

        const candidate = parsed as Partial<StoredSchoolAccess>
        if (typeof candidate.schoolID !== "string" || !/^\d+$/.test(candidate.schoolID)) return null
        if (!candidate.access || typeof candidate.access !== "object") return null
        if (typeof candidate.access.owner !== "boolean" || !Array.isArray(candidate.access.roles)) return null

        return {schoolID: candidate.schoolID, access: candidate.access}
    } catch {
        return null
    }
}

export function rememberActiveSchool(schoolID: string) {
    if (!/^\d+$/.test(schoolID)) return
    sessionStorage.setItem(ACTIVE_SCHOOL_STORAGE_KEY, schoolID)
    window.dispatchEvent(new Event(ACTIVE_SCHOOL_EVENT))
}

export function rememberSchoolAccess(schoolID: string, access: SchoolAccess) {
    if (!/^\d+$/.test(schoolID)) return

    sessionStorage.setItem(ACTIVE_SCHOOL_STORAGE_KEY, schoolID)
    sessionStorage.setItem(ACTIVE_SCHOOL_ACCESS_KEY, JSON.stringify({schoolID, access}))
    window.dispatchEvent(new Event(ACTIVE_SCHOOL_EVENT))
}

export function rememberCurrentSchoolAccess(access: SchoolAccess) {
    const schoolID = getActiveSchoolSnapshot()
    if (schoolID) rememberSchoolAccess(schoolID, access)
}
