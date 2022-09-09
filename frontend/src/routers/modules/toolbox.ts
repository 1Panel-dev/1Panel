import { Layout } from '@/routers/constant';

const toolBoxRouter = {
    sort: 9,
    path: '/toolbox',
    component: Layout,
    redirect: '/toolbox',
    meta: {
        icon: 'p-toolbox',
        title: 'menu.toolbox',
    },
    children: [
        {
            path: '/toolbox',
            name: 'ToolBox',
            component: () => import('@/views/toolbox/index.vue'),
            meta: {},
        },
    ],
};

export default toolBoxRouter;
