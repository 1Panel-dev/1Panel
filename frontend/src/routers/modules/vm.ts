import { Layout } from '@/routers/constant';
import vmRouter from '@/routers/vm';

const virtualMachineRouter = {
    sort: 6.5,
    path: '/vm',
    name: 'VirtualMachine',
    component: Layout,
    redirect: '/vm/overview',
    meta: {
        icon: 'p-xuniji',
        title: 'xpack.vm.vm',
        requiresAuth: true,
        permission: 'xpack_vm_view',
    },
    children: [vmRouter],
};

export default virtualMachineRouter;
