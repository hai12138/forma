import type { Metadata } from 'next';
import './globals.css';
import './product.css';
export const metadata: Metadata = {
  metadataBase: new URL(
    'https://forma-business-agent-studio.cocoa-maple-0505.chatgpt.site',
  ),
  title: 'Forma · Business-to-Agent Platform',
  description: '从业务事实到可信赖的 Agent 应用。企业工程平台交互原型。',
  icons: { icon: '/favicon.svg' },
  openGraph: {
    title: 'Forma · Business-to-Agent Platform',
    description: '从业务事实到可信赖的 Agent 应用。企业工程平台交互原型。',
    images: [
      {
        url: 'https://forma-business-agent-studio.cocoa-maple-0505.chatgpt.site/og.png',
        width: 1731,
        height: 909,
        alt: 'Forma Business-to-Agent Platform',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Forma · Business-to-Agent Platform',
    description: '从业务事实到可信赖的 Agent 应用。',
    images: [
      'https://forma-business-agent-studio.cocoa-maple-0505.chatgpt.site/og.png',
    ],
  },
};
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
