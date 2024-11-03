import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primaryBackground: "var(--primary-background)",
        secondaryBackground: "var(--secondary-background)",
        secondaryBackgroundVariant: "var(--secondary-background-variant)",
        heroTextColor: "var(--hero-text-color)",
        borderColor: "var(--border-color)",
      },
    },
  },
  plugins: [],
};
export default config;
