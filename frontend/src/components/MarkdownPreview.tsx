import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeKatex from 'rehype-katex';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import 'katex/dist/katex.min.css';

interface MarkdownPreviewProps {
  content: string;
  compact?: boolean;
}

export function MarkdownPreview({ content, compact = false }: MarkdownPreviewProps) {
  return (
    <article className={compact ? 'markdown-body markdown-body-compact' : 'markdown-body'}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, [remarkMath, { singleDollarTextMath: true }]]}
        rehypePlugins={[rehypeKatex, rehypeHighlight]}
      >
        {content || '\n'}
      </ReactMarkdown>
    </article>
  );
}
