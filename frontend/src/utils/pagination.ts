import type { ReqPage, ResPage, ResultData } from '@/api/interface';

export async function readPages<T>(search: (page: ReqPage) => Promise<ResultData<ResPage<T>>>) {
    const items: T[] = [];
    for (let page = 1; ; page++) {
        const res = await search({ page, pageSize: 100 });
        const batch = res.data.items || [];
        items.push(...batch);
        if (items.length >= res.data.total || batch.length === 0) {
            return items;
        }
    }
}
