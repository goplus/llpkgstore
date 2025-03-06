import { useRef } from 'react';
import { CSSTransition } from 'react-transition-group';

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    children?: React.ReactNode;
    title?: string;
    closeButtonClassName?: string;
    contentClassName?: string;
    overlayClassName?: string;
}

const Modal: React.FC<ModalProps> = ({
    isOpen,
    onClose,
    children,
    title,
    closeButtonClassName = '',
    contentClassName = '',
    overlayClassName = '',
}) => {
    const overlayRef = useRef<HTMLDivElement>(null);
    const contentRef = useRef<HTMLDivElement>(null);

    const handleOverlayClick = (e: React.MouseEvent) => {
        if (e.target === e.currentTarget) onClose();
    };

    return (
        <>
            <CSSTransition
                in={isOpen}
                timeout={300}
                classNames={{
                    enter: 'opacity-0',
                    enterActive: 'opacity-100 transition-opacity duration-300',
                    exitActive: 'opacity-0 transition-opacity duration-300',
                }}
                unmountOnExit
                nodeRef={overlayRef}
            >
                <div
                    ref={overlayRef}
                    className={`fixed inset-0 z-50 bg-black/50 ${overlayClassName}`}
                    onClick={handleOverlayClick}
                />
            </CSSTransition>

            <CSSTransition
                in={isOpen}
                timeout={300}
                classNames={{
                    enter: 'opacity-0 scale-95',
                    enterActive:
                        'opacity-100 scale-100 transition-all duration-300',
                    exitActive:
                        'opacity-0 scale-95 transition-all duration-300',
                }}
                unmountOnExit
                onExited={onClose}
                nodeRef={contentRef}
            >
                <div
                    ref={contentRef}
                    className={`fixed top-1/2 left-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 transform rounded-xl bg-white p-6 shadow-lg ${contentClassName}`}
                    onClick={(e) => e.stopPropagation()}
                >
                    <div className="mb-4 flex items-center justify-between">
                        {title && (
                            <h3 className="text-xl font-semibold">{title}</h3>
                        )}
                        <button
                            className={`absolute top-4 right-4 cursor-pointer rounded-xl p-2 text-gray-500 transition-all duration-300 hover:bg-gray-100 hover:text-gray-700 ${closeButtonClassName}`}
                            onClick={onClose}
                        >
                            <svg
                                className="h-6 w-6"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </div>
                    <div className="max-h-[80vh] overflow-y-auto">
                        {children}
                    </div>
                </div>
            </CSSTransition>
        </>
    );
};

export default Modal;
