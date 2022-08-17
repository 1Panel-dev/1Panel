import { Layout } from '@/routers/constant';

const toolBoxRouter = {
    sort: 7,
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
            meta: {
                keepAlive: true,
            },
        },
    ],
};

export default toolBoxRouter;
