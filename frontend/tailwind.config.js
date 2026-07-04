/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./components/**/*.{ts,tsx}", "./pages/**/*.{ts,tsx}", "./router/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          DEFAULT: "#0a0a0b",
          soft: "#111113",
          muted: "#1a1a1f",
        },
        signal: {
          DEFAULT: "#00ff41",
          dim: "#2e8b57",
          glow: "#3dff7a",
        },
        fog: {
          DEFAULT: "#8b8b96",
          light: "#b4b4be",
        },
        line: "#2a2a32",
      },
      fontFamily: {
        display: ['"Syne"', "system-ui", "sans-serif"],
        body: ['"DM Sans"', "system-ui", "sans-serif"],
        mono: ['"IBM Plex Mono"', "monospace"],
      },
      animation: {
        "fade-up": "fadeUp 0.7s ease-out forwards",
        "pulse-signal": "pulseSignal 3s ease-in-out infinite",
      },
      keyframes: {
        fadeUp: {
          "0%": { opacity: "0", transform: "translateY(16px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        pulseSignal: {
          "0%, 100%": { opacity: "0.4" },
          "50%": { opacity: "1" },
        },
      },
    },
  },
  plugins: [],
};