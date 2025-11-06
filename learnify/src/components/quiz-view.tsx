"use client";

import { useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { QuizBundle } from "@/types/course";

interface QuizViewProps {
  quiz: QuizBundle;
}

export function QuizView({ quiz }: QuizViewProps) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [feedback, setFeedback] = useState<Record<string, string>>({});

  const score = useMemo(() => {
    const correct = quiz.questions.filter((question) => answers[question.id] === question.answer).length;
    return `${correct} / ${quiz.questions.length}`;
  }, [answers, quiz.questions]);

  function handleSelect(questionId: string, optionId: string) {
    const question = quiz.questions.find((item) => item.id === questionId);
    if (!question) return;

    setAnswers((prev) => ({ ...prev, [questionId]: optionId }));
    if (question.answer === optionId) {
      setFeedback((prev) => ({ ...prev, [questionId]: "✅ 回答正确！" }));
    } else {
      const hint = question.hint ?? "再读一遍原文中的相关段落，尝试理解关键词。";
      setFeedback((prev) => ({ ...prev, [questionId]: `❌ 不正确。提示：${hint}` }));
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-semibold">选择题挑战</h2>
          <p className="text-sm text-muted-foreground">点击选项后立即查看提示或解析</p>
        </div>
        <Badge variant="secondary">当前得分 {score}</Badge>
      </div>

      <div className="grid gap-4">
        {quiz.questions.map((question, index) => (
          <Card key={question.id} className="border-muted bg-secondary/40">
            <CardHeader>
              <Badge variant="outline" className="w-fit">
                题目 {index + 1}
              </Badge>
              <CardTitle className="text-lg">{question.prompt}</CardTitle>
              <CardDescription>{question.context}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid gap-2">
                {question.options.map((option) => {
                  const isSelected = answers[question.id] === option.id;
                  const isCorrect = question.answer === option.id;
                  return (
                    <Button
                      key={option.id}
                      variant={isSelected ? (isCorrect ? "secondary" : "destructive") : "outline"}
                      className="justify-start"
                      onClick={() => handleSelect(question.id, option.id)}
                    >
                      <span className="mr-3 inline-flex h-6 w-6 items-center justify-center rounded-full border border-border text-xs">
                        {String.fromCharCode(65 + option.index)}
                      </span>
                      {option.text}
                    </Button>
                  );
                })}
              </div>
              {feedback[question.id] ? (
                <p className="text-sm text-muted-foreground">{feedback[question.id]}</p>
              ) : null}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
