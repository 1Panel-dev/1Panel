"use client";

import Image from "next/image";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { CertificatePayload } from "@/types/course";

interface CertificatePreviewProps {
  certificate: CertificatePayload;
}

export function CertificatePreview({ certificate }: CertificatePreviewProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState<string | null>(certificate.imageUrl ?? null);

  async function regenerate() {
    setIsLoading(true);
    try {
      const response = await fetch("/api/generate-certificate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          learner: certificate.learner,
          courseTitle: certificate.courseTitle,
          achievements: certificate.achievements,
        }),
      });

      if (!response.ok) {
        throw new Error("证书生成失败");
      }

      const data = (await response.json()) as { imageUrl: string };
      setImageUrl(data.imageUrl);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Card className="border-muted bg-secondary/30">
      <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <CardTitle>AI 学习证书</CardTitle>
          <CardDescription>
            完成练习后，你可保存证书截图或继续生成新的视觉风格。
          </CardDescription>
        </div>
        <Button onClick={regenerate} disabled={isLoading} variant="outline">
          {isLoading ? "生成中..." : "重新生成证书"}
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <ul className="space-y-2 text-sm text-muted-foreground">
            <li>学习者：{certificate.learner}</li>
            <li>课程主题：{certificate.courseTitle}</li>
            <li>达成目标：{certificate.achievements.join("，")}</li>
          </ul>
          <div className="relative min-h-[240px] overflow-hidden rounded-xl border border-dashed border-primary/50 bg-gradient-to-br from-primary/20 via-background to-accent/20">
            {imageUrl ? (
              <Image src={imageUrl} alt="学习证书" fill className="object-cover" />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-sm text-muted-foreground">
                证书图片将在生成后展示
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
