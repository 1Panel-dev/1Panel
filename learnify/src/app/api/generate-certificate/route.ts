import { NextResponse } from "next/server";
import { z } from "zod";

const schema = z.object({
  learner: z.string().min(1),
  courseTitle: z.string().min(1),
  achievements: z.array(z.string().min(1)).min(1),
});

export async function POST(request: Request) {
  try {
    if (!process.env.OPENAI_API_KEY) {
      return NextResponse.json(
        { message: "缺少 OPENAI_API_KEY" },
        { status: 500 }
      );
    }

    const body = await request.json();
    const { learner, courseTitle, achievements } = schema.parse(body);

    const prompt = `设计一张商务暗黑风格的学习证书，主标题为“随身学 Learnify 证书”，学习者姓名 ${learner}，课程主题 ${courseTitle}，成就包括：${achievements.join("，")}。`;

    const response = await fetch("https://api.openai.com/v1/images/generations", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${process.env.OPENAI_API_KEY}`,
      },
      body: JSON.stringify({
        model: "gpt-image-1",
        prompt,
        size: "1024x1024",
        style: "vivid",
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`证书生成失败: ${errorText}`);
    }

    const data = await response.json();
    const imageUrl: string | undefined = data?.data?.[0]?.url;

    if (!imageUrl) {
      throw new Error("未返回证书图片");
    }

    return NextResponse.json({ imageUrl });
  } catch (error) {
    console.error(error);
    if (error instanceof z.ZodError) {
      return NextResponse.json({ message: "参数无效" }, { status: 422 });
    }
    return NextResponse.json({ message: error instanceof Error ? error.message : "证书生成异常" }, { status: 500 });
  }
}
