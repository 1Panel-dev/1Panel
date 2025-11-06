import type { Metadata } from "next";
import "../styles/globals.css";
import { cn } from "@/lib/utils";
import { SonnerToaster } from "@/components/sonner-toaster";

export const metadata: Metadata = {
  title: "随身学 Learnify",
  description: "将任何文章一键转换为交互式微课程",
  icons: {
    icon: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body
        className={cn(
          "min-h-screen bg-background font-sans text-foreground antialiased",
          "bg-[radial-gradient(circle_at_top,_rgba(79,70,229,0.3),_transparent_55%)]"
        )}
      >
        {children}
        <div
          className="pointer-events-none fixed inset-0 bg-grid-glow opacity-40 mix-blend-soft-light"
          aria-hidden
        />
        <SonnerToaster />
      </body>
    </html>
  );
}
