"use client"

import {useMemo, useRef, useState} from "react"
import {
    AlertCircle, ChevronLeft, ChevronRight, Download, FileSpreadsheet, FileUp,
    Keyboard, Plus, StickyNote, Trash2, Upload, UsersRound,
} from "lucide-react"
import {toast} from "sonner"
import {Button} from "@/components/ui/button"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle,
} from "@/components/ui/dialog"
import {Input} from "@/components/ui/input"
import {Switch} from "@/components/ui/switch"
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table"
import {Tabs, TabsList, TabsPanel, TabsTab} from "@/components/ui/tabs"
import {Textarea} from "@/components/ui/textarea"
import {useLocale} from "@/i18n/provider"
import {parseStudentImportFile, type StudentImportParseError, type StudentImportRow} from "@/lib/student_import"
import {handleImportStudents} from "./actions"
import {studentErrorKeys} from "./student_action_errors"

type ImportMode = "manual" | "file"
type RowField = keyof StudentImportRow
type RowError = "name" | "lastName" | "dateOfBirth" | "email" | "phone" | "notes" | "duplicateEmail"

interface EditableStudentRow extends StudentImportRow {
    id: string
}

const pageSize = 25
const maxFileSize = 5 * 1024 * 1024

function blankRow(id: string): EditableStudentRow {
    return {id, name: "", lastName: "", dateOfBirth: "", email: "", phone: "", notes: ""}
}

function initialManualRows() {
    return [blankRow("manual-1"), blankRow("manual-2"), blankRow("manual-3")]
}

function rowHasContent(row: EditableStudentRow) {
    return [row.name, row.lastName, row.dateOfBirth, row.email, row.phone, row.notes]
        .some((value) => value.trim())
}

function validDateOnly(value: string) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false
    const date = new Date(`${value}T00:00:00Z`)
    return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value
}

function validateRow(row: EditableStudentRow, duplicateEmails: Set<string>): RowError[] {
    const errors: RowError[] = []
    const name = row.name.trim()
    const lastName = row.lastName.trim()
    const email = row.email.trim().toLocaleLowerCase("en-US")
    const phone = row.phone.trim()
    const notes = row.notes.trim()

    if (name.length < 3 || name.length > 32) errors.push("name")
    if (lastName.length < 3 || lastName.length > 32) errors.push("lastName")
    if (!validDateOnly(row.dateOfBirth.trim())) errors.push("dateOfBirth")
    if (email.length < 5 || email.length > 254 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) errors.push("email")
    if (phone && (phone.length < 3 || phone.length > 32)) errors.push("phone")
    if (notes.length > 2048) errors.push("notes")
    if (duplicateEmails.has(email)) errors.push("duplicateEmail")
    return errors
}

function parseErrorKey(error: StudentImportParseError) {
    return `staff.students.import.fileError.${error}` as const
}

export default function BulkStudentImport({open, schoolID, onOpenChange, onImported}: {
    open: boolean
    schoolID: string
    onOpenChange: (open: boolean) => void
    onImported: () => void
}) {
    const {t} = useLocale()
    const inputRef = useRef<HTMLInputElement>(null)
    const nextManualID = useRef(4)
    const [mode, setMode] = useState<ImportMode>("manual")
    const [manualRows, setManualRows] = useState<EditableStudentRow[]>(initialManualRows)
    const [fileRows, setFileRows] = useState<EditableStudentRow[]>([])
    const [fileName, setFileName] = useState<string | null>(null)
    const [manualPage, setManualPage] = useState(1)
    const [filePage, setFilePage] = useState(1)
    const [enableAccounts, setEnableAccounts] = useState(false)
    const [importing, setImporting] = useState(false)
    const [serverInvalidRowID, setServerInvalidRowID] = useState<string | null>(null)
    const [serverConflictRowIDs, setServerConflictRowIDs] = useState<string[]>([])
    const [noteEditorRowID, setNoteEditorRowID] = useState<string | null>(null)
    const [noteDraft, setNoteDraft] = useState("")

    const sourceRows = mode === "manual" ? manualRows : fileRows
    const activeRows = useMemo(
        () => mode === "manual" ? sourceRows.filter(rowHasContent) : sourceRows,
        [mode, sourceRows],
    )
    const duplicateEmails = useMemo(() => {
        const counts = new Map<string, number>()
        activeRows.forEach((row) => {
            const email = row.email.trim().toLocaleLowerCase("en-US")
            if (email) counts.set(email, (counts.get(email) ?? 0) + 1)
        })
        return new Set([...counts].flatMap(([email, count]) => count > 1 ? [email] : []))
    }, [activeRows])
    const errorsByRow = useMemo(() => new Map(activeRows.map((row) => [row.id, validateRow(row, duplicateEmails)])), [activeRows, duplicateEmails])
    const invalidRows = activeRows.filter((row) => (errorsByRow.get(row.id)?.length ?? 0) > 0)
    const currentPage = mode === "manual" ? manualPage : filePage
    const setCurrentPage = mode === "manual" ? setManualPage : setFilePage
    const pageCount = Math.max(1, Math.ceil(sourceRows.length / pageSize))
    const safePage = Math.min(currentPage, pageCount)
    const visibleRows = sourceRows.slice((safePage - 1) * pageSize, safePage * pageSize)
    const activeServerInvalidRowID = sourceRows.some((row) => row.id === serverInvalidRowID) ? serverInvalidRowID : null
    const sourceRowIDs = new Set(sourceRows.map((row) => row.id))
    const activeServerConflictRowIDs = new Set(serverConflictRowIDs.filter((id) => sourceRowIDs.has(id)))
    const invalidRowIDs = new Set(invalidRows.map((row) => row.id))
    const serverIssueIDs = new Set([
        ...(activeServerInvalidRowID ? [activeServerInvalidRowID] : []),
        ...activeServerConflictRowIDs,
    ])
    const issueCount = invalidRows.length + [...serverIssueIDs].filter((id) => !invalidRowIDs.has(id)).length
    const canImport = activeRows.length > 0
        && activeRows.length <= 10_000
        && issueCount === 0
        && !importing

    function reset() {
        setMode("manual")
        setManualRows(initialManualRows())
        setFileRows([])
        setFileName(null)
        setManualPage(1)
        setFilePage(1)
        setEnableAccounts(false)
        setServerInvalidRowID(null)
        setServerConflictRowIDs([])
        setNoteEditorRowID(null)
        setNoteDraft("")
        nextManualID.current = 4
        if (inputRef.current) inputRef.current.value = ""
    }

    function close() {
        if (importing) return
        reset()
        onOpenChange(false)
    }

    function changeMode(value: string | number) {
        if (value !== "manual" && value !== "file") return
        setMode(value)
    }

    function updateRow(id: string, field: RowField, value: string) {
        const update = (rows: EditableStudentRow[]) => rows.map((row) => row.id === id ? {...row, [field]: value} : row)
        if (mode === "manual") setManualRows(update)
        else setFileRows(update)
        if (serverInvalidRowID === id) setServerInvalidRowID(null)
        if (field === "email" && serverConflictRowIDs.includes(id)) {
            setServerConflictRowIDs((ids) => ids.filter((rowID) => rowID !== id))
        }
    }

    function removeRow(id: string) {
        if (mode === "manual") {
            setManualRows((rows) => {
                const remaining = rows.filter((row) => row.id !== id)
                return remaining.length ? remaining : [blankRow(`manual-${nextManualID.current++}`)]
            })
        } else {
            setFileRows((rows) => rows.filter((row) => row.id !== id))
        }
        if (serverInvalidRowID === id) setServerInvalidRowID(null)
        if (serverConflictRowIDs.includes(id)) setServerConflictRowIDs((ids) => ids.filter((rowID) => rowID !== id))
    }

    function addManualRow() {
        const next = blankRow(`manual-${nextManualID.current++}`)
        setManualRows((rows) => [...rows, next])
        setManualPage(Math.ceil((manualRows.length + 1) / pageSize))
    }

    function openNoteEditor(row: EditableStudentRow) {
        setNoteEditorRowID(row.id)
        setNoteDraft(row.notes)
    }

    function closeNoteEditor() {
        setNoteEditorRowID(null)
        setNoteDraft("")
    }

    function saveNote() {
        if (!noteEditorRowID) return
        updateRow(noteEditorRowID, "notes", noteDraft)
        closeNoteEditor()
    }

    async function loadFile(file: File) {
        const extensionAllowed = /\.(csv|tsv|txt)$/i.test(file.name)
        if (!extensionAllowed) return toast.error(t("staff.students.import.fileTypeError"))
        if (file.size > maxFileSize) return toast.error(t("staff.students.import.fileSizeError"))

        let contents: string
        try {
            contents = await file.text()
        } catch {
            return toast.error(t("staff.students.import.fileReadError"))
        }

        const result = parseStudentImportFile(contents)
        if (!result.ok) {
            const headerLabels: Record<string, string> = {
                name: t("staff.students.firstName"),
                lastName: t("staff.students.lastName"),
                dateOfBirth: t("staff.students.dateOfBirth"),
                email: t("staff.students.email"),
            }
            const details = result.error === "missing_headers" && result.missingHeaders
                ? ` ${t("staff.students.import.missingHeaders", {
                    headers: result.missingHeaders.map((header) => headerLabels[header] ?? header).join(", "),
                })}`
                : ""
            return toast.error(`${t(parseErrorKey(result.error))}${details}`)
        }

        setFileRows(result.rows.map((row, index) => ({...row, id: `file-${index + 1}`})))
        setFileName(file.name)
        setFilePage(1)
        setServerInvalidRowID(null)
        setServerConflictRowIDs([])
        toast.success(t("staff.students.import.fileLoaded", {count: result.rows.length}))
    }

    function downloadTemplate() {
        const template = "firstName,lastName,dateOfBirth,email,phone,notes\r\nAlex,Johnson,2010-04-23,alex.johnson@example.com,+1 555 0100,Optional note\r\n"
        const url = URL.createObjectURL(new Blob([template], {type: "text/csv;charset=utf-8"}))
        const anchor = document.createElement("a")
        anchor.href = url
        anchor.download = "edulink-student-import-template.csv"
        anchor.click()
        URL.revokeObjectURL(url)
    }

    async function importStudents() {
        if (!canImport) return
        const rows = activeRows
        setImporting(true)
        const result = await handleImportStudents(schoolID, rows.map((row) => ({
            name: row.name.trim(),
            lastName: row.lastName.trim(),
            dateOfBirth: row.dateOfBirth.trim(),
            email: row.email.trim().toLocaleLowerCase("en-US"),
            phone: row.phone.trim() || null,
            notes: row.notes.trim() || null,
        })), enableAccounts)
        setImporting(false)

        if (!result.ok) {
            if ("conflictingEmails" in result) {
                const conflictingEmails = new Set(result.conflictingEmails)
                const conflictRows = rows.filter((row) => conflictingEmails.has(row.email.trim().toLocaleLowerCase("en-US")))
                setServerConflictRowIDs(conflictRows.map((row) => row.id))
                const firstConflictRow = conflictRows[0]
                if (firstConflictRow) {
                    const sourceRowIndex = sourceRows.findIndex((row) => row.id === firstConflictRow.id)
                    setCurrentPage(Math.floor(sourceRowIndex / pageSize) + 1)
                }
                return toast.error(result.conflictingEmails.length === 1
                    ? t("staff.students.import.emailConflict", {email: result.conflictingEmails[0]})
                    : t("staff.students.import.emailConflicts", {count: result.conflictingEmails.length}))
            }
            if ("invalidRow" in result) {
                const invalidRow = rows[result.invalidRow]
                if (invalidRow) {
                    setServerInvalidRowID(invalidRow.id)
                    const sourceRowIndex = sourceRows.findIndex((row) => row.id === invalidRow.id)
                    const page = Math.floor(sourceRowIndex / pageSize) + 1
                    setCurrentPage(page)
                    return toast.error(t("staff.students.import.serverInvalidRow", {row: sourceRowIndex + 1}))
                }
            }
            return toast.error(t(studentErrorKeys[result.code]))
        }

        toast.success(t("staff.students.import.success", {count: rows.length}))
        reset()
        onOpenChange(false)
        onImported()
    }

    return (
        <Dialog open={open} onOpenChange={(nextOpen) => { if (!nextOpen) close() }}>
            <DialogPopup className="sm:max-w-6xl">
                <DialogHeader>
                    <DialogTitle>{t("staff.students.import.title")}</DialogTitle>
                    <DialogDescription>{t("staff.students.import.description")}</DialogDescription>
                </DialogHeader>
                <DialogPanel className="space-y-5">
                    <Tabs value={mode} onValueChange={changeMode}>
                        <TabsList className="w-full sm:w-fit" aria-label={t("staff.students.import.modeLabel")}>
                            <TabsTab className="flex-1 sm:flex-none" value="manual"><Keyboard /> {t("staff.students.import.manualTab")}</TabsTab>
                            <TabsTab className="flex-1 sm:flex-none" value="file"><FileSpreadsheet /> {t("staff.students.import.fileTab")}</TabsTab>
                        </TabsList>

                        <TabsPanel className="pt-3" value="manual">
                            <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
                                <div><p className="font-medium">{t("staff.students.import.manualTitle")}</p><p className="text-sm text-muted-foreground">{t("staff.students.import.manualDescription")}</p></div>
                                <Button size="sm" variant="outline" onClick={addManualRow}><Plus /> {t("staff.students.import.addRow")}</Button>
                            </div>
                            <StudentImportTable
                                rows={visibleRows}
                                rowOffset={(safePage - 1) * pageSize}
                                errorsByRow={errorsByRow}
                                serverInvalidRowID={activeServerInvalidRowID}
                                serverConflictRowIDs={activeServerConflictRowIDs}
                                disabled={importing}
                                onChange={updateRow}
                                onEditNotes={openNoteEditor}
                                onRemove={removeRow}
                            />
                        </TabsPanel>

                        <TabsPanel className="space-y-4 pt-3" value="file">
                            <div
                                className="flex min-h-44 flex-col items-center justify-center rounded-2xl border border-dashed bg-muted/20 px-5 py-8 text-center transition-colors hover:bg-muted/35"
                                onDragOver={(event) => event.preventDefault()}
                                onDrop={(event) => {
                                    event.preventDefault()
                                    const file = event.dataTransfer.files[0]
                                    if (file) void loadFile(file)
                                }}
                            >
                                <span className="mb-3 flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary"><FileUp /></span>
                                <p className="font-medium">{t("staff.students.import.dropTitle")}</p>
                                <p className="mt-1 max-w-lg text-sm text-muted-foreground">{t("staff.students.import.dropDescription")}</p>
                                <div className="mt-4 flex flex-wrap justify-center gap-2">
                                    <Button size="sm" onClick={() => inputRef.current?.click()}><Upload /> {t("staff.students.import.chooseFile")}</Button>
                                    <Button size="sm" variant="outline" onClick={downloadTemplate}><Download /> {t("staff.students.import.downloadTemplate")}</Button>
                                </div>
                                <input
                                    className="sr-only"
                                    ref={inputRef}
                                    type="file"
                                    accept=".csv,.tsv,.txt,text/csv,text/tab-separated-values,text/plain"
                                    onChange={(event) => {
                                        const file = event.target.files?.[0]
                                        if (file) void loadFile(file)
                                        event.target.value = ""
                                    }}
                                />
                            </div>
                            {fileName && (
                                <div className="flex flex-wrap items-center justify-between gap-2 rounded-xl border bg-card px-4 py-3 text-sm">
                                    <span className="flex min-w-0 items-center gap-2"><FileSpreadsheet className="size-4 shrink-0 text-primary" /><span className="truncate font-medium">{fileName}</span></span>
                                    <span className="text-muted-foreground">{t("staff.students.import.rowsLoaded", {count: fileRows.length})}</span>
                                </div>
                            )}
                            {fileRows.length > 0 && (
                                <StudentImportTable
                                    rows={visibleRows}
                                    rowOffset={(safePage - 1) * pageSize}
                                    errorsByRow={errorsByRow}
                                    serverInvalidRowID={activeServerInvalidRowID}
                                    serverConflictRowIDs={activeServerConflictRowIDs}
                                    disabled={importing}
                                    onChange={updateRow}
                                    onEditNotes={openNoteEditor}
                                    onRemove={removeRow}
                                />
                            )}
                        </TabsPanel>
                    </Tabs>

                    {sourceRows.length > pageSize && (
                        <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
                            <p className="text-sm text-muted-foreground">{t("staff.students.import.page", {page: safePage, total: pageCount})}</p>
                            <div className="flex gap-2">
                                <Button size="sm" variant="outline" disabled={safePage === 1 || importing} onClick={() => setCurrentPage(Math.max(1, safePage - 1))}><ChevronLeft /> {t("staff.students.import.previousPage")}</Button>
                                <Button size="sm" variant="outline" disabled={safePage === pageCount || importing} onClick={() => setCurrentPage(Math.min(pageCount, safePage + 1))}>{t("staff.students.import.nextPage")} <ChevronRight /></Button>
                            </div>
                        </div>
                    )}

                    <div className="grid gap-3 rounded-2xl border bg-muted/20 p-4 sm:grid-cols-[1fr_auto] sm:items-center">
                        <div>
                            <p className="font-medium">{t("staff.students.import.enableAccounts")}</p>
                            <p className="text-sm text-muted-foreground">{t("staff.students.import.enableAccountsDescription")}</p>
                        </div>
                        <Switch checked={enableAccounts} onCheckedChange={setEnableAccounts} disabled={importing} aria-label={t("staff.students.import.enableAccounts")} />
                    </div>

                    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-xl border px-4 py-3 text-sm">
                        <span className="flex items-center gap-2 font-medium"><UsersRound className="size-4 text-primary" /> {t("staff.students.import.readyCount", {count: activeRows.length})}</span>
                        {issueCount > 0 && <span className="flex items-center gap-1.5 text-destructive"><AlertCircle className="size-4" /> {t("staff.students.import.issueCount", {count: issueCount})}</span>}
                    </div>
                </DialogPanel>
                <DialogFooter>
                    <DialogClose render={<Button variant="outline" disabled={importing} />}>{t("common.cancel")}</DialogClose>
                    <Button loading={importing} disabled={!canImport} onClick={importStudents}><Upload /> {t("staff.students.import.submit", {count: activeRows.length})}</Button>
                </DialogFooter>

                <Dialog open={noteEditorRowID !== null} onOpenChange={(nextOpen) => { if (!nextOpen) closeNoteEditor() }}>
                    <DialogPopup className="sm:max-w-xl">
                        <DialogHeader>
                            <DialogTitle>{t("staff.students.import.noteTitle")}</DialogTitle>
                            <DialogDescription>{t("staff.students.import.noteDescription")}</DialogDescription>
                        </DialogHeader>
                        <DialogPanel className="space-y-2">
                            <Textarea
                                className="min-h-48 resize-y"
                                value={noteDraft}
                                maxLength={2048}
                                disabled={importing}
                                autoFocus
                                placeholder={t("staff.students.notesPlaceholder")}
                                onChange={(event) => setNoteDraft(event.target.value)}
                            />
                            <p className="text-right text-xs tabular-nums text-muted-foreground">{noteDraft.length}/2048</p>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" disabled={importing} />}>{t("common.cancel")}</DialogClose>
                            <Button disabled={importing} onClick={saveNote}>{t("staff.students.import.saveNote")}</Button>
                        </DialogFooter>
                    </DialogPopup>
                </Dialog>
            </DialogPopup>
        </Dialog>
    )
}

function StudentImportTable({rows, rowOffset, errorsByRow, serverInvalidRowID, serverConflictRowIDs, disabled, onChange, onEditNotes, onRemove}: {
    rows: EditableStudentRow[]
    rowOffset: number
    errorsByRow: Map<string, RowError[]>
    serverInvalidRowID: string | null
    serverConflictRowIDs: Set<string>
    disabled: boolean
    onChange: (id: string, field: RowField, value: string) => void
    onEditNotes: (row: EditableStudentRow) => void
    onRemove: (id: string) => void
}) {
    const {t} = useLocale()
    const errorLabels: Record<RowError, string> = {
        name: t("staff.students.import.error.name"),
        lastName: t("staff.students.import.error.lastName"),
        dateOfBirth: t("staff.students.import.error.dateOfBirth"),
        email: t("staff.students.import.error.email"),
        phone: t("staff.students.import.error.phone"),
        notes: t("staff.students.import.error.notes"),
        duplicateEmail: t("staff.students.import.error.duplicateEmail"),
    }

    return (
        <div className="overflow-hidden rounded-xl border">
            <Table>
                <TableHeader>
                    <TableRow className="bg-muted/35">
                        <TableHead className="w-12">#</TableHead>
                        <TableHead>{t("staff.students.firstName")}</TableHead>
                        <TableHead>{t("staff.students.lastName")}</TableHead>
                        <TableHead>{t("staff.students.dateOfBirth")}</TableHead>
                        <TableHead>{t("staff.students.email")}</TableHead>
                        <TableHead>{t("staff.students.phone")}</TableHead>
                        <TableHead>{t("staff.students.notes")}</TableHead>
                        <TableHead className="w-12"><span className="sr-only">{t("staff.students.import.actions")}</span></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {rows.map((row, index) => {
                        const rowNumber = rowOffset + index + 1
                        const errors = errorsByRow.get(row.id) ?? []
                        const serverConflict = serverConflictRowIDs.has(row.id)
                        const hasError = errors.length > 0 || serverInvalidRowID === row.id || serverConflict
                        return (
                            <StudentImportTableRow
                                row={row}
                                rowNumber={rowNumber}
                                errors={errors}
                                hasError={hasError}
                                serverConflict={serverConflict}
                                errorLabels={errorLabels}
                                disabled={disabled}
                                onChange={onChange}
                                onEditNotes={onEditNotes}
                                onRemove={onRemove}
                                key={row.id}
                            />
                        )
                    })}
                </TableBody>
            </Table>
        </div>
    )
}

function StudentImportTableRow({row, rowNumber, errors, hasError, serverConflict, errorLabels, disabled, onChange, onEditNotes, onRemove}: {
    row: EditableStudentRow
    rowNumber: number
    errors: RowError[]
    hasError: boolean
    serverConflict: boolean
    errorLabels: Record<RowError, string>
    disabled: boolean
    onChange: (id: string, field: RowField, value: string) => void
    onEditNotes: (row: EditableStudentRow) => void
    onRemove: (id: string) => void
}) {
    const {t} = useLocale()
    const fieldError = (field: RowField) => errors.includes(field as RowError)
    const fieldLabels: Record<RowField, string> = {
        name: t("staff.students.firstName"),
        lastName: t("staff.students.lastName"),
        dateOfBirth: t("staff.students.dateOfBirth"),
        email: t("staff.students.email"),
        phone: t("staff.students.phone"),
        notes: t("staff.students.notes"),
    }
    const input = (field: RowField, type = "text", className = "min-w-40") => (
        <Input
            className={className}
            type={type}
            value={row[field]}
            disabled={disabled}
            aria-invalid={fieldError(field)}
            aria-label={`${fieldLabels[field]} ${rowNumber}`}
            onChange={(event) => onChange(row.id, field, event.target.value)}
        />
    )

    return (
        <>
            <TableRow className={hasError ? "bg-destructive/4" : undefined}>
                <TableCell className="text-xs tabular-nums text-muted-foreground">{rowNumber}</TableCell>
                <TableCell>{input("name")}</TableCell>
                <TableCell>{input("lastName")}</TableCell>
                <TableCell>{input("dateOfBirth", "date", "min-w-40")}</TableCell>
                <TableCell>{input("email", "email", "min-w-64")}</TableCell>
                <TableCell>{input("phone", "tel")}</TableCell>
                <TableCell className="min-w-64">
                    <div className="flex items-center gap-2">
                        <p className="min-w-0 flex-1 truncate text-xs text-muted-foreground" title={row.notes || undefined}>
                            {row.notes.trim() || t("staff.students.import.noNote")}
                        </p>
                        <Button size="xs" variant="outline" disabled={disabled} onClick={() => onEditNotes(row)}>
                            <StickyNote /> {t(row.notes.trim() ? "staff.students.import.editNote" : "staff.students.import.addNote")}
                        </Button>
                    </div>
                </TableCell>
                <TableCell>
                    <Button size="icon-xs" variant="ghost" disabled={disabled} aria-label={t("staff.students.import.removeRow", {row: rowNumber})} onClick={() => onRemove(row.id)}><Trash2 /></Button>
                </TableCell>
            </TableRow>
            {hasError && (
                <TableRow className="bg-destructive/4">
                    <TableCell className="whitespace-normal py-2 text-xs text-destructive" colSpan={8}>
                        <span className="flex items-start gap-1.5">
                            <AlertCircle className="mt-0.5 size-3.5 shrink-0" />
                            {errors.length
                                ? errors.map((error) => errorLabels[error]).join(" ")
                                : serverConflict
                                    ? t("staff.students.import.error.existingEmail", {email: row.email.trim()})
                                    : t("staff.students.import.error.serverRow")}
                        </span>
                    </TableCell>
                </TableRow>
            )}
        </>
    )
}
