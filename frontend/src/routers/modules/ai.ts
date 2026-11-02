import { Layout } from '@/routers/constant';
import { GlobalStore } from '@/store';

const aiRouter = {
    sort: 4,
    path: '/ai',
    name: 'AI-Menu',
    component: Layout,
    redirect: '/ai/model',
    meta: {
        icon: 'p-jiqiren2',
        title: 'menu.aiTools',
    },
    children: [
        {
            path: '/ai/model',
            name: 'AIModel',
            redirect: '/ai/model/files',
            meta: {
                icon: 'p-moxing-menu',
                title: 'aiTools.model.model',
                activeMenu: '/ai/model',
                permission: 'ai_model_view',
            },
        },
        {
            path: '/ai/model/files',
            hidden: true,
            name: 'ModelFiles',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'aiTools.model.files',
                activeMenu: '/ai/model',
                permission: 'ai_model_view',
            },
        },
        {
            path: '/ai/model/account',
            hidden: true,
            name: 'ModelServices',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'aiTools.model.services',
                activeMenu: '/ai/model',
                permission: 'ai_model_view',
            },
        },
        {
            path: '/ai/model/local',
            hidden: true,
            name: 'LocalModel',
            component: () => import('@/views/ai/model/index.vue'),
            meta: {
                title: 'aiTools.model.deployments',
                activeMenu: '/ai/model',
                permission: 'ai_model_view',
            },
        },
        {
            path: '/ai/agents/agent',
            name: 'Agents',
            component: () => import('@/views/ai/agents/agent/index.vue'),
            meta: {
                icon: 'p-jiqiren2',
                title: 'aiTools.agents.agent',
                permission: 'ai_agent_view',
            },
        },
        {
            path: '/ai/mcp',
            name: 'MCPServer',
            redirect: () => (GlobalStore().hasPermission('ai_mcp_view') ? '/ai/mcp/server' : '/ai/mcp/1panel'),
            meta: {
                icon: 'p-mcp-menu',
                title: 'menu.mcp',
                permission: ['ai_mcp_view', 'mcp_access_view'],
            },
        },
        {
            path: '/ai/mcp/server',
            name: 'MCPServerList',
            hidden: true,
            component: () => import('@/views/ai/mcp/server/index.vue'),
            meta: {
                title: 'menu.mcp',
                activeMenu: '/ai/mcp',
                permission: 'ai_mcp_view',
            },
        },
        {
            path: '/ai/mcp/1panel',
            name: 'PanelMCP',
            hidden: true,
            component: () => import('@/views/ai/mcp/panel/index.vue'),
            meta: {
                title: 'panelMCP.title',
                activeMenu: '/ai/mcp',
                permission: 'mcp_access_view',
            },
        },
        {
            path: '/ai/gpu/current',
            name: 'GPU',
            component: () => import('@/views/ai/gpu/current/index.vue'),
            meta: {
                icon: 'p-gpu-menu',
                title: 'aiTools.gpu.gpu',
                activeMenu: '/ai/gpu/current',
                permission: 'ai_gpu_view',
            },
        },
        {
            path: '/ai/gpu/history',
            name: 'GPUHistory',
            component: () => import('@/views/ai/gpu/history/index.vue'),
            hidden: true,
            meta: {
                title: 'aiTools.gpu.history',
                activeMenu: '/ai/gpu/current',
                permission: 'ai_gpu_view',
            },
        },
    ],
};

export default aiRouter;
