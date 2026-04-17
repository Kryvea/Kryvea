module.exports = {
  darkMode: "class",
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      transitionProperty: {
        position: "right, left, top, bottom, margin, padding",
        textColor: "color",
      },
    },
  },
  plugins: [require("@tailwindcss/container-queries")],
};
