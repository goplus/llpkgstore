import { useState } from 'react';
import { VersionData } from '../../tools/parser/types';
import Modal from '../modal';
import { versionParser } from '../../tools/parser/parser';

interface DetailModalProps {
    modalOpen: boolean;
    setModalOpen: (open: boolean) => void;
    name?: string;
    data?: VersionData;
}

const DetailModal: React.FC<DetailModalProps> = ({
    modalOpen,
    setModalOpen,
    name,
    data,
}) => {
    const [originVersion, setOriginVersion] = useState('');
    const [convertVersion, setConvertVersion] = useState('');
    return (
        <Modal
            isOpen={modalOpen}
            onClose={() => setModalOpen(false)}
            title={name}
        >
            {/* <Details /> */}
            {data && name && (
                <>
                    <div className="my-2 flex flex-row justify-around text-base">
                        <div className="flex items-center gap-2 rounded-3xl border border-gray-300 px-2 py-1">
                            <input
                                onChange={(e) =>
                                    setOriginVersion(e.target.value)
                                }
                                className="px-1 focus-visible:outline-0"
                                placeholder="Original version"
                            />
                        </div>
                        <div className="flex items-center gap-2 rounded-3xl border border-gray-300 px-2 py-1">
                            <input
                                onChange={(e) =>
                                    setConvertVersion(e.target.value)
                                }
                                className="px-1 focus-visible:outline-0"
                                placeholder="Converted version"
                            />
                        </div>
                    </div>
                    {versionParser(
                        data[name],
                        originVersion,
                        convertVersion,
                        1,
                        10,
                    ).map((ver, index) => {
                        return (
                            <div
                                key={index}
                                className="flex flex-row items-center gap-4"
                            >
                                <span className="min-w-16 text-left text-lg font-bold">
                                    {ver.original}
                                </span>
                                <span>{ver.converted.join(' / ')}</span>
                            </div>
                        );
                    })}
                </>
            )}
        </Modal>
    );
};

export default DetailModal;
