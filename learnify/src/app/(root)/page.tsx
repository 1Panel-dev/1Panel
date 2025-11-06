"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Award, Loader2, ListChecks, Sparkles, Upload, Wand2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { CertificatePreview } from "@/components/certificate-preview";
import { QuizView } from "@/components/quiz-view";
import type { CourseBundle } from "@/types/course";

const PLACEHOLDER = "粘贴一段你想深入理解的文章、新闻或笔记...";

export default function HomePage() {
  const [sourceText, setSourceText] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [bundle, setBundle] = useState<CourseBundle | null>(null);

  async function handleGenerate() {
    if (!sourceText.trim()) {
      toast.warning("请先粘贴需要学习的原文");
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch("/api/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: sourceText }),
      });

      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(error.message ?? "生成内容失败");
      }

      const data: CourseBundle = await response.json();
      setBundle(data);
      toast.success("已生成 Learnify 微课程");
    } catch (error) {
      console.error(error);
      toast.error(error instanceof Error ? error.message : "服务暂时不可用，请稍后再试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-8 px-4 py-10 sm:px-6">
      <header className="flex flex-col gap-3 text-center">
        <Badge className="mx-auto bg-primary/20 text-primary" variant="outline">
          随身学 Learnify Beta
        </Badge>
        <h1 className="font-display text-3xl font-semibold sm:text-4xl">
          将任何文章一键变成可练习的微课程
        </h1>
        <p className="mx-auto max-w-2xl text-sm text-muted-foreground sm:text-base">
          利用 AI 自动提炼 3 个核心知识点，生成 5 道交互式选择题和开放式练习题，
          并输出一张学习证书，帮助你快速检验是否真正读懂。
        </p>
      </header>

      <Card className="border-muted bg-secondary/40 backdrop-blur">
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1 text-left">
            <CardTitle>粘贴原文</CardTitle>
            <CardDescription>我们将自动识别文章重点并设计练习题</CardDescription>
          </div>
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" onClick={() => setSourceText("")}>
              <Upload className="mr-2 h-4 w-4" />
              清空
            </Button>
            <Button onClick={handleGenerate} disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  生成中...
                </>
              ) : (
                <>
                  <Wand2 className="mr-2 h-4 w-4" />
                  一键生成课程
                </>
              )}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <Textarea
            value={sourceText}
            onChange={(event) => setSourceText(event.target.value)}
            placeholder={PLACEHOLDER}
            className="min-h-[220px] bg-secondary/60"
          />
          <p className="text-xs text-muted-foreground">
            Learnify 使用 OpenRouter 与 GPT Image API，数据默认不会被用作模型训练。
          </p>
        </CardContent>
      </Card>

      {bundle ? (
        <Tabs defaultValue="insights" className="grid gap-6">
          <TabsList className="mx-auto grid w-full max-w-xl grid-cols-3 bg-secondary/60">
            <TabsTrigger value="insights" className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" />
              知识点
            </TabsTrigger>
            <TabsTrigger value="quiz" className="flex items-center gap-2">
              <ListChecks className="h-4 w-4" />
              习题
            </TabsTrigger>
            <TabsTrigger value="certificate" className="flex items-center gap-2">
              <Award className="h-4 w-4" />
              证书
            </TabsTrigger>
          </TabsList>
          <TabsContent value="insights" className="space-y-4">
            <section className="grid gap-4 sm:grid-cols-3">
              {bundle.insights.map((item, index) => (
                <Card key={index} className="border-muted bg-secondary/40">
                  <CardHeader>
                    <Badge variant="secondary" className="w-fit">核心知识点 {index + 1}</Badge>
                    <CardTitle className="text-lg">{item.title}</CardTitle>
                    <CardDescription>{item.summary}</CardDescription>
                  </CardHeader>
                </Card>
              ))}
            </section>
            <Card className="border-dashed border-muted bg-secondary/40">
              <CardHeader>
                <CardTitle>开放式练习题</CardTitle>
                <CardDescription>{bundle.openPractice.prompt}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2 text-sm text-muted-foreground">
                <p>回答建议：</p>
                <ul className="list-disc space-y-1 pl-5">
                  {bundle.openPractice.guidance.map((tip, index) => (
                    <li key={index}>{tip}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="quiz">
            <QuizView quiz={bundle.quiz} />
          </TabsContent>

          <TabsContent value="certificate">
            <CertificatePreview certificate={bundle.certificate} />
          </TabsContent>
        </Tabs>
      ) : (
        <Card className="border-muted bg-secondary/30">
          <CardHeader className="items-center text-center">
            <Sparkles className="mb-4 h-10 w-10 text-primary" />
            <CardTitle>你的微课程将展示在这里</CardTitle>
            <CardDescription>粘贴文本并点击“一键生成课程”开始体验</CardDescription>
          </CardHeader>
        </Card>
      )}
    </main>
  );
}
