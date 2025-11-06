export interface Insight {
  title: string;
  summary: string;
}

export interface QuizOption {
  id: string;
  index: number;
  text: string;
}

export interface QuizQuestion {
  id: string;
  prompt: string;
  context: string;
  options: QuizOption[];
  answer: string;
  hint?: string;
}

export interface QuizBundle {
  questions: QuizQuestion[];
}

export interface OpenPractice {
  prompt: string;
  guidance: string[];
}

export interface CertificatePayload {
  learner: string;
  courseTitle: string;
  achievements: string[];
  imageUrl?: string;
}

export interface CourseBundle {
  insights: Insight[];
  quiz: QuizBundle;
  openPractice: OpenPractice;
  certificate: CertificatePayload;
  createdAt: string;
  source: string;
}
