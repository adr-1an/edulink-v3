"use client"

import {useEffect, useRef, useState} from "react"
import {
    ClipboardPaste, Copy, ExternalLink, Image as ImageIcon, Link2, Scissors, TextSelect,
} from "lucide-react"
import {useLocale} from "@/i18n/provider"

type EditableElement = HTMLInputElement | HTMLTextAreaElement | HTMLElement

type MenuTarget =
    | {kind: "link"; url: string}
    | {kind: "image"; url: string}
    | {kind: "editable"; element: EditableElement; canEdit: boolean; canCopy: boolean}
    | {kind: "selection"; text: string}

interface MenuState {
    x: number
    y: number
    target: MenuTarget
}

interface MenuAction {
    label: string
    icon: typeof Copy
    run: () => void | Promise<void>
}

function selectedEditableText(element: EditableElement) {
    if (element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement) {
        const start = element.selectionStart ?? 0
        const end = element.selectionEnd ?? 0
        return element.value.slice(start, end)
    }
    return window.getSelection()?.toString() ?? ""
}

async function copyText(text: string) {
    if (!text) return
    if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
        return
    }

    const textarea = document.createElement("textarea")
    textarea.value = text
    textarea.style.position = "fixed"
    textarea.style.opacity = "0"
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand("copy")
    textarea.remove()
}

function dispatchInput(element: HTMLInputElement | HTMLTextAreaElement, inputType: string, data: string | null = null) {
    element.dispatchEvent(new InputEvent("input", {bubbles: true, inputType, data}))
}

async function cutFrom(element: EditableElement) {
    if (element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement) {
        const start = element.selectionStart ?? 0
        const end = element.selectionEnd ?? 0
        await copyText(element.value.slice(start, end))
        element.setRangeText("", start, end, "end")
        dispatchInput(element, "deleteByCut")
        return
    }
    element.focus()
    document.execCommand("cut")
}

async function pasteInto(element: EditableElement) {
    const text = await navigator.clipboard.readText()
    element.focus()
    if (element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement) {
        const start = element.selectionStart ?? element.value.length
        const end = element.selectionEnd ?? start
        element.setRangeText(text, start, end, "end")
        dispatchInput(element, "insertFromPaste", text)
        return
    }
    document.execCommand("insertText", false, text)
}

function selectAll(element: EditableElement) {
    element.focus()
    if (element instanceof HTMLInputElement || element instanceof HTMLTextAreaElement) {
        element.select()
        return
    }

    const range = document.createRange()
    range.selectNodeContents(element)
    const selection = window.getSelection()
    selection?.removeAllRanges()
    selection?.addRange(range)
}

function getMenuTarget(eventTarget: EventTarget | null): MenuTarget | null {
    if (!(eventTarget instanceof Element)) return null

    const link = eventTarget.closest<HTMLAnchorElement>("a[href]")
    if (link) return {kind: "link", url: link.href}

    const image = eventTarget.closest<HTMLImageElement>("img[src]")
    if (image) return {kind: "image", url: image.currentSrc || image.src}

    const editable = eventTarget.closest<HTMLElement>("input, textarea, [contenteditable='true']")
    if (editable) {
        const isFormField = editable instanceof HTMLInputElement || editable instanceof HTMLTextAreaElement
        const canEdit = isFormField ? !editable.disabled && !editable.readOnly : true
        const isPassword = editable instanceof HTMLInputElement && editable.type === "password"
        return {
            kind: "editable",
            element: editable,
            canEdit,
            canCopy: !isPassword && Boolean(selectedEditableText(editable)),
        }
    }

    const selection = window.getSelection()
    const text = selection?.toString().trim() ?? ""
    if (text && selection?.rangeCount) {
        const range = selection.getRangeAt(0)
        if (range.intersectsNode(eventTarget)) return {kind: "selection", text}
    }

    return null
}

export default function GlobalContextMenu() {
    const {t} = useLocale()
    const [menu, setMenu] = useState<MenuState | null>(null)
    const menuRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        function handleContextMenu(event: MouseEvent) {
            event.preventDefault()
            const target = getMenuTarget(event.target)
            if (!target) {
                setMenu(null)
                return
            }

            const width = 224
            const estimatedHeight = target.kind === "editable" ? 140 : 108
            setMenu({
                x: Math.min(event.clientX, window.innerWidth - width - 8),
                y: Math.min(event.clientY, window.innerHeight - estimatedHeight - 8),
                target,
            })
        }

        function closeMenu(event?: Event) {
            if (event?.target instanceof Node && menuRef.current?.contains(event.target)) return
            setMenu(null)
        }

        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") closeMenu()
        }

        document.addEventListener("contextmenu", handleContextMenu, true)
        document.addEventListener("pointerdown", closeMenu)
        window.addEventListener("blur", closeMenu)
        window.addEventListener("resize", closeMenu)
        window.addEventListener("scroll", closeMenu, true)
        document.addEventListener("keydown", handleKeyDown)
        return () => {
            document.removeEventListener("contextmenu", handleContextMenu, true)
            document.removeEventListener("pointerdown", closeMenu)
            window.removeEventListener("blur", closeMenu)
            window.removeEventListener("resize", closeMenu)
            window.removeEventListener("scroll", closeMenu, true)
            document.removeEventListener("keydown", handleKeyDown)
        }
    }, [])

    useEffect(() => {
        menuRef.current?.querySelector<HTMLButtonElement>("button")?.focus()
    }, [menu])

    if (!menu) return null

    const actions: MenuAction[] = []
    if (menu.target.kind === "link") {
        const target = menu.target
        actions.push(
            {label: t("context.openLink"), icon: Link2, run: () => window.location.assign(target.url)},
            {label: t("context.openNewTab"), icon: ExternalLink, run: () => { window.open(target.url, "_blank", "noopener,noreferrer") }},
            {label: t("context.copyLink"), icon: Copy, run: () => copyText(target.url)},
        )
    } else if (menu.target.kind === "image") {
        const target = menu.target
        actions.push(
            {label: t("context.openImageNewTab"), icon: ImageIcon, run: () => { window.open(target.url, "_blank", "noopener,noreferrer") }},
            {label: t("context.copyImage"), icon: Copy, run: () => copyText(target.url)},
        )
    } else if (menu.target.kind === "selection") {
        const target = menu.target
        actions.push({label: t("context.copy"), icon: Copy, run: () => copyText(target.text)})
    } else {
        const target = menu.target
        if (target.canCopy) {
            actions.push({label: t("context.copy"), icon: Copy, run: () => copyText(selectedEditableText(target.element))})
            if (target.canEdit) actions.push({label: t("context.cut"), icon: Scissors, run: () => cutFrom(target.element)})
        }
        if (target.canEdit && window.isSecureContext) {
            actions.push({label: t("context.paste"), icon: ClipboardPaste, run: () => pasteInto(target.element)})
        }
        actions.push({label: t("context.selectAll"), icon: TextSelect, run: () => selectAll(target.element)})
    }

    function runAction(action: MenuAction) {
        setMenu(null)
        void Promise.resolve(action.run()).catch(() => undefined)
    }

    function handleMenuKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
        if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
        event.preventDefault()
        const items = [...event.currentTarget.querySelectorAll<HTMLButtonElement>("button")]
        if (!items.length) return
        const current = items.indexOf(document.activeElement as HTMLButtonElement)
        if (event.key === "Home") return items[0].focus()
        if (event.key === "End") return items.at(-1)?.focus()
        const offset = event.key === "ArrowDown" ? 1 : -1
        items[(current + offset + items.length) % items.length].focus()
    }

    return (
        <div
            ref={menuRef}
            role="menu"
            aria-label={t("context.label")}
            className="fixed z-[100] w-56 rounded-lg border bg-popover p-1 text-popover-foreground shadow-lg outline-none"
            style={{left: Math.max(8, menu.x), top: Math.max(8, menu.y)}}
            onContextMenu={(event) => event.preventDefault()}
            onKeyDown={handleMenuKeyDown}
        >
            {actions.map((action) => {
                const Icon = action.icon
                return (
                    <button
                        key={action.label}
                        type="button"
                        role="menuitem"
                        className="flex min-h-8 w-full cursor-default items-center gap-2 rounded-sm px-2 py-1 text-left text-sm outline-none hover:bg-accent focus-visible:bg-accent"
                        onClick={() => runAction(action)}
                    >
                        <Icon className="size-4 opacity-80" />
                        {action.label}
                    </button>
                )
            })}
        </div>
    )
}
