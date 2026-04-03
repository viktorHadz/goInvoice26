import type { ActiveEditorNode } from '@/components/editor/invBookTypes'

export function resolveEditorExportRevisionNo(
    activeNode: ActiveEditorNode,
    activeRevisionNo: number,
): number {
    if (activeNode?.type === 'revision') {
        return activeNode.revisionNo
    }

    return activeRevisionNo > 0 ? activeRevisionNo : 1
}
