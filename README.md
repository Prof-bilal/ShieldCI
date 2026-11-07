# ShieldCI Landing Page

A modern, responsive landing page for ShieldCI - Zero-Trust CI/CD Security powered by Warden.

## 🚀 Features

- **Next.js 14** - Modern React framework
- **Tailwind CSS** - Utility-first CSS
- **Framer Motion** - Smooth animations
- **Responsive Design** - Works on all devices
- **SEO Optimized** - Built-in SEO support

## 📁 Project Structure

```
landing/
├── pages/
│   └── index.js          # Main landing page
├── public/
│   └── favicon.svg       # Favicon
├── components/           # Reusable components (to be added)
├── styles/
│   └── globals.css       # Global styles
├── package.json
├── tailwind.config.js
├── next.config.js
├── postcss.config.js
├── tsconfig.json
├── .gitignore
└── README.md
```

## 🛠️ Getting Started

### Prerequisites

- Node.js 18.17 or later
- npm or yarn

### Installation

1. Clone the repository:
```bash
cd /home/abullah/Projects/ShieldCI/landing
```

2. Install dependencies:
```bash
npm install
# or
yarn install
```

3. Run the development server:
```bash
npm run dev
# or
yarn dev
```

4. Open [http://localhost:3000](http://localhost:3000) in your browser.

## 🎨 Customization

### Update Content

Edit `pages/index.js` to modify:
- Hero section text and buttons
- Features list
- Pricing tiers
- Testimonials
- FAQ items
- Company information

### Update Styles

Edit `styles/globals.css` to modify:
- Colors
- Fonts
- Custom animations
- Background patterns

### Update Configuration

- **Tailwind CSS**: Edit `tailwind.config.js`
- **Next.js**: Edit `next.config.js`

## 📦 Deploy

### Vercel (Recommended)

1. Push your code to a Git repository (GitHub, GitLab, Bitbucket)
2. Import the project in [Vercel](https://vercel.com)
3. Deploy!

### Other Platforms

- Netlify
- AWS Amplify
- Heroku
- Any Node.js hosting

## 🔧 Configuration

### Environment Variables

Create a `.env.local` file for local development:

```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

## 🎯 Sections Included

1. **Header** - Navigation bar with logo and CTA
2. **Hero** - Main value proposition with email signup
3. **Stats** - Industry statistics
4. **Features** - 6 key features with icons
5. **Recent Attacks** - Showcase of recent breaches
6. **How It Works** - 3-step process
7. **Competitive Comparison** - Feature comparison table
8. **Use Cases** - 4 use case cards
9. **Pricing** - 3 pricing tiers
10. **Testimonials** - Customer testimonials
11. **FAQ** - Frequently asked questions
12. **CTA** - Final call-to-action
13. **Footer** - Links and social media

## 💡 Tips

### Add Real Testimonials

Replace the placeholder testimonials in `pages/index.js` with real customer quotes.

### Connect Email Form

Update the `handleSubmit` function in `pages/index.js` to connect to your email service (Mailchimp, SendGrid, etc.).

### Add Analytics

Add Google Analytics or other tracking:

```bash
npm install react-ga
```

Then update `pages/_app.js` (create if doesn't exist).

### Add More Pages

Create additional pages in the `pages/` directory:
- `pages/about.js`
- `pages/docs.js`
- `pages/blog.js`
- `pages/contact.js`

## 📊 Performance

This landing page is optimized for:
- Fast load times
- Mobile responsiveness
- SEO
- Accessibility

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## 📜 License

MIT License

---

Built with ❤️ using Next.js and Tailwind CSS
