import { FileText } from 'lucide-react';

export function EmptyEditor() {
  return (
    <section className="empty-editor">
      <FileText size={42} />
      <h2>没有打开的文件</h2>
    </section>
  );
}
