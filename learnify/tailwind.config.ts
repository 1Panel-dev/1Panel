import type { Config } from "tailwindcss";
import { fontFamily } from "tailwindcss/defaultTheme";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./src/app/**/*.{ts,tsx}",
    "./src/components/**/*.{ts,tsx}",
    "./src/lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#0b0d17",
        foreground: "#f4f6fb",
        muted: {
          DEFAULT: "#1b1f2a",
          foreground: "#9aa4b2",
        },
        primary: {
          DEFAULT: "#4f46e5",
          foreground: "#f9fafb",
        },
        secondary: {
          DEFAULT: "#1f2937",
          foreground: "#e5e7eb",
        },
        accent: {
          DEFAULT: "#14b8a6",
          foreground: "#052e27",
        },
        destructive: {
          DEFAULT: "#ef4444",
          foreground: "#fef2f2",
        },
        border: "#1f2432",
        input: "#1f2432",
        ring: "#4338ca",
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      fontFamily: {
        sans: ["Inter", ...fontFamily.sans],
        display: ["Sora", ...fontFamily.sans],
      },
      backgroundImage: {
        "grid-glow": "linear-gradient(135deg, rgba(79,70,229,0.15), rgba(20,184,166,0.12))",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};

export default config;
