import { MarkdownPreview } from './MarkdownPreview';

interface PreviewPaneProps {
  content: string;
  fileName: string;
}

export function PreviewPane({ content, fileName }: PreviewPaneProps) {
  return (
    <section className="preview-pane" aria-label="Markdown preview">
      <div className="pane-titlebar">
        <span>{fileName}</span>
      </div>
      <div className="preview-scroll">
        <MarkdownPreview content={content} />
      </div>
    </section>
  );
}
