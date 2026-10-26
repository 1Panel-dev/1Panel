// Wait for every request, keeping at most four deletions in flight. Return only
// failed items so a retry cannot delete an already-removed item a second time.
export async function deleteFilesWithLimit<T>(
    items: readonly T[],
    remove: (item: T) => Promise<unknown>,
): Promise<T[]> {
    let next = 0;
    const failed = new Set<number>();
    const worker = async () => {
        while (next < items.length) {
            const index = next++;
            try {
                await remove(items[index]);
            } catch {
                failed.add(index);
            }
        }
    };
    await Promise.all(Array.from({ length: Math.min(4, items.length) }, worker));
    return items.filter((_, index) => failed.has(index));
}
