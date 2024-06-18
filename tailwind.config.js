/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./frontend/**/*.html",
    "./frontend/**/*.gohtml",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}

