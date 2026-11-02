const vmRouter = {
    sort: 550,
    path: '/vm',
    name: 'VirtualMachine',
    component: () => import('@/views/vm/index.vue'),
    redirect: '/vm/overview',
    meta: {
        icon: 'p-disk-menu',
        title: 'xpack.vm.title',
        requiresAuth: true,
        permission: 'xpack_vm_view',
    },
    children: [
        {
            path: 'overview',
            name: 'VirtualMachineOverview',
            component: () => import('@/views/vm/overview/index.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.overview',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
        {
            path: 'list',
            name: 'VirtualMachineList',
            component: () => import('@/views/vm/list/index.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.title',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
        {
            path: 'isos',
            name: 'VirtualMachineISOs',
            component: () => import('@/views/vm/iso/iso.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.iso',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
        {
            path: 'templates',
            name: 'VirtualMachineTemplates',
            component: () => import('@/views/vm/template/index.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.template',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
        {
            path: 'networks',
            name: 'VirtualMachineNetworks',
            component: () => import('@/views/vm/network/index.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.network',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
        {
            path: 'storage-pools',
            name: 'VirtualMachineStoragePools',
            component: () => import('@/views/vm/storage/index.vue'),
            hidden: true,
            meta: {
                title: 'xpack.vm.storagePool',
                activeMenu: '/vm',
                requiresAuth: true,
                permission: 'xpack_vm_view',
            },
        },
    ],
};

export default vmRouter;
