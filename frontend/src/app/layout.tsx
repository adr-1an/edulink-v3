import type { Metadata } from "next";
import "./globals.css";
import NextTopLoader from "nextjs-toploader"
import { cn } from "@/lib/utils";
import GlobalContextMenu from "@/components/app/global_context_menu";
import {cookies} from "next/headers";
import {ThemeProvider, type Theme} from "@/components/app/theme_provider";
import {LocaleProvider} from "@/i18n/provider";
import {getLocale} from "@/i18n/server";

export const metadata: Metadata = {
    title: "EduLink",
    description: "EduLink",
};

export default async function RootLayout({
                                       children,
                                   }: Readonly<{
    children: React.ReactNode;
}>) {
    const savedTheme = (await cookies()).get("theme")?.value
    const theme: Theme = savedTheme === "dark" ? "dark" : "light"
    const locale = await getLocale()

    return (
        <html
            lang={locale}
            className={cn("h-full font-sans antialiased", theme === "dark" && "dark")}
            suppressHydrationWarning
        >
        <body className="min-h-full">
        <LocaleProvider locale={locale}>
        <ThemeProvider initialTheme={theme}>
        <NextTopLoader color="#0071e3" showSpinner={false} />
            {children}
            <GlobalContextMenu />
        </ThemeProvider>
        </LocaleProvider>
        </body>
        </html>
    );
}
