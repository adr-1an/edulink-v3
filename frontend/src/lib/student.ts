export interface StudentDraft {
    name: string
    lastName: string
    dateOfBirth: string
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    accountActive: boolean
    password: string
}

export interface StudentInput {
    name: string
    lastName: string
    dob: string | null
    email: string
    phone: string
    notes: string
    accountEnabled: boolean
    accountActive?: boolean
    password: string
}

export const emptyStudentDraft: StudentDraft = {
    name: "",
    lastName: "",
    dateOfBirth: "",
    email: "",
    phone: "",
    notes: "",
    accountEnabled: false,
    accountActive: false,
    password: "",
}

export function studentDraftToInput(draft: StudentDraft): StudentInput {
    return {
        name: draft.name.trim(),
        lastName: draft.lastName.trim(),
        dob: draft.dateOfBirth || null,
        email: draft.email.trim().toLocaleLowerCase(),
        phone: draft.phone.trim(),
        notes: draft.notes.trim(),
        accountEnabled: draft.accountEnabled,
        ...(draft.accountEnabled ? {accountActive: draft.accountActive} : {}),
        password: draft.password,
    }
}
