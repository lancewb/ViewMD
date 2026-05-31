import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import remarkGfm from 'remark-gfm';

interface MarkdownPreviewProps {
  content: string;
  compact?: boolean;
}

export function MarkdownPreview({ content, compact = false }: MarkdownPreviewProps) {
  return (
    <article className={compact ? 'markdown-body markdown-body-compact' : 'markdown-body'}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
        {content || '\n'}
      </ReactMarkdown>
    </article>
  );
}
