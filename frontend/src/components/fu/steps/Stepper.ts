import i18n from '@/lang';

import type { StepOptions, StepperOptions } from './types';

export class Step implements StepOptions {
    id?: string;
    index: number;
    beforeActive?: Function;
    beforeLeave?: Function;
    title?: string;
    description?: string;
    icon?: string;
    status?: string;

    constructor(options?: StepOptions) {
        const stepOptions = options ?? ({ index: 0 } as StepOptions);
        this.id = stepOptions.id;
        this.index = stepOptions.index;
        this.beforeActive = stepOptions.beforeActive;
        this.beforeLeave = stepOptions.beforeLeave;
        this.title = stepOptions.title;
        this.description = stepOptions.description;
        this.icon = stepOptions.icon;
        this.status = stepOptions.status;
    }
}

export class Stepper implements StepperOptions {
    steps: Step[];
    index: number;
    activeSet: Set<number>;
    isLoading: boolean;
    onCancelButtonText: string;
    onFinishButtonText: string;
    prevButtonText: string;
    nextButtonText: string;
    buttonSize: string;
    footerAlign: string;
    showCancel: boolean;
    beforeActive?: Function;
    beforeLeave?: Function;
    height?: string | number;

    constructor(options?: Partial<StepperOptions>) {
        this.steps = options?.steps?.map((step) => new Step(step)) || [];
        this.index = options?.index ?? 0;
        this.activeSet = new Set<number>();
        this.isLoading = options?.isLoading ?? false;
        this.onCancelButtonText = options?.onCancelButtonText ?? i18n.global.t('fu.steps.cancel');
        this.onFinishButtonText = options?.onFinishButtonText ?? i18n.global.t('fu.steps.finish');
        this.prevButtonText = options?.prevButtonText ?? i18n.global.t('fu.steps.prev');
        this.nextButtonText = options?.nextButtonText ?? i18n.global.t('fu.steps.next');
        this.buttonSize = options?.buttonSize ?? 'default';
        this.footerAlign = options?.footerAlign ?? 'flex';
        this.showCancel = options?.showCancel ?? false;
        this.beforeActive = options?.beforeActive;
        this.beforeLeave = options?.beforeLeave;
        this.height = options?.height;
    }

    isFirst(index: number) {
        return index === 0;
    }

    isLast(index: number) {
        return index === this.steps.length - 1;
    }

    isActive(index: number) {
        return this.activeSet.has(index);
    }

    isCurrent(index: number) {
        return this.index === index;
    }

    async active(index: number) {
        const isValid = index >= 0 && index < this.steps.length && this.index !== index;
        const forward = index > this.index;
        if (!isValid) {
            return;
        }
        if ((await this.executeBeforeLeave(this.index, forward)) === false) {
            return;
        }
        if ((await this.executeBeforeActive(index, forward)) === false) {
            return;
        }
        this.index = index;
        this.activeSet.add(index);
    }

    next() {
        if (!this.isLast(this.index)) {
            return this.active(this.index + 1);
        }
    }

    prev() {
        if (!this.isFirst(this.index)) {
            return this.active(this.index - 1);
        }
    }

    getStep(index: number) {
        return this.steps[index];
    }

    executeBeforeLeave(index: number, forward: boolean) {
        const step = this.getStep(index);
        if (step?.beforeLeave) {
            return step.beforeLeave(step, forward);
        }
        return this.beforeLeave?.(step, forward);
    }

    executeBeforeActive(index: number, forward: boolean) {
        const step = this.getStep(index);
        if (step?.beforeActive) {
            return step.beforeActive(step, forward);
        }
        return this.beforeActive?.(step, forward);
    }
}
