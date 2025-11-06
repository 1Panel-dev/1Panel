import type { CourseBundle } from "@/types/course";

const OPENROUTER_URL = "https://openrouter.ai/api/v1/chat/completions";

export interface GenerateCourseParams {
  text: string;
  model?: string;
}

export async function requestCourseBundle({ text, model = "deepseek/deepseek-chat" }: GenerateCourseParams) {
  if (!process.env.OPENROUTER_API_KEY) {
    throw new Error("缺少 OPENROUTER_API_KEY 环境变量");
  }

  const payload = {
    model,
    temperature: 0.7,
    response_format: { type: "json_object" },
    messages: [
      {
        role: "system",
        content:
          "你是 Learnify 的课程设计助手。请从输入文本中提炼知识点并输出 JSON，字段包括 insights(数组)、quiz.questions(数组)、openPractice 对象、certificate 对象。",
      },
      {
        role: "user",
        content: `请基于以下文本生成 Learnify 微课程：\n${text}\n\n返回 JSON，格式如下：\n{\n  "insights": [{"title": string, "summary": string}],\n  "quiz": {\n    "questions": [{\n      "id": string,\n      "prompt": string,\n      "context": string,\n      "options": [{"id": string, "index": number, "text": string}],\n      "answer": string,\n      "hint": string\n    }]\n  },\n  "openPractice": {"prompt": string, "guidance": string[]},\n  "certificate": {"learner": string, "courseTitle": string, "achievements": string[]},\n  "createdAt": string\n}\n要求：\n- insights 必须是 3 条。\n- quiz.questions 必须是 5 道，选项数量为 4。\n- certificate.learner 默认填写 \"Learnify Explorer\"。` ,
      },
    ],
  };

  const response = await fetch(OPENROUTER_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${process.env.OPENROUTER_API_KEY}`,
      "HTTP-Referer": process.env.OPENROUTER_APP_URL ?? "https://learnify.ai",
      "X-Title": "Learnify Course Generator",
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`OpenRouter 请求失败: ${errorText}`);
  }

  const data = await response.json();
  const content = data?.choices?.[0]?.message?.content;
  if (!content) {
    throw new Error("OpenRouter 返回为空");
  }

  const parsed = JSON.parse(content) as CourseBundle;
  parsed.source = text;
  if (!parsed.createdAt) {
    parsed.createdAt = new Date().toISOString();
  }
  return parsed;
}
