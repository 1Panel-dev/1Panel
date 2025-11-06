import { NextResponse } from "next/server";
import { z } from "zod";
import { db } from "@/lib/db";
import { courses } from "@/lib/schema";
import { requestCourseBundle } from "@/lib/openrouter";

const requestSchema = z.object({
  text: z.string().min(40, "请提供不少于 40 个字符的学习材料")
});

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const { text } = requestSchema.parse(body);

    const bundle = await requestCourseBundle({ text });

    if (db) {
      await db.insert(courses).values({ source: text, bundle });
    }

    return NextResponse.json(bundle);
  } catch (error) {
    console.error(error);
    if (error instanceof z.ZodError) {
      return NextResponse.json({ message: error.issues[0]?.message ?? "参数错误" }, { status: 422 });
    }
    return NextResponse.json({ message: error instanceof Error ? error.message : "服务异常" }, { status: 500 });
  }
}
