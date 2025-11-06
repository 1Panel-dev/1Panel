import { json, pgTable, text, timestamp, uuid } from "drizzle-orm/pg-core";

export const courses = pgTable("courses", {
  id: uuid("id").defaultRandom().primaryKey(),
  createdAt: timestamp("created_at", { withTimezone: true }).defaultNow(),
  source: text("source"),
  bundle: json("bundle"),
});

export type InsertCourse = typeof courses.$inferInsert;
