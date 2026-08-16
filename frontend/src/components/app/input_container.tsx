import {ReactNode} from "react"

export default function InputContainer({ children, className }: { children: ReactNode, className?: string }) {
    return (
        <div className={`flex flex-col mb-4 ${className}`}>
            {children}
        </div>
    )
}