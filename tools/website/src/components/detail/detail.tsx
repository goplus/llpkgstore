import { useEffect, useState } from 'react';
import { VersionData } from '../../tools/parser/types';
import { versionParser } from '../../tools/parser/parser';
import Modal from '../modal';
import Title from './title';
import VersionItem from './items';

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
    const [mappedVersion, setMappedVersion] = useState('');
    const [version, setVersion] = useState('latest');

    useEffect(() => {
        setVersion('latest');
    }, [name]);
    return (
        <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)}>
            <div className="flex h-full flex-col">
                <Title name={name} version={version} />
                <div className="relative h-full overflow-auto px-4 pb-3">
                    {data && name && (
                        <>
                            <div className="sticky top-1 mt-5 mb-2 flex flex-row justify-around text-base">
                                <div className="mb-4 flex items-center gap-2 rounded-lg border border-gray-300 bg-white/70 px-2 backdrop-blur-lg">
                                    <input
                                        onChange={(e) =>
                                            setOriginVersion(e.target.value)
                                        }
                                        className="border-r border-gray-300 px-1 py-2 focus-visible:outline-0"
                                        placeholder="Original version"
                                    />
                                    <input
                                        onChange={(e) =>
                                            setMappedVersion(e.target.value)
                                        }
                                        className="px-1 py-2 focus-visible:outline-0"
                                        placeholder="Mapped version"
                                    />
                                </div>
                            </div>
                            <div className="flex h-full flex-col gap-2">
                                {versionParser(
                                    data[name],
                                    originVersion,
                                    mappedVersion,
                                    0,
                                    10,
                                ).map((ver, index) => {
                                    return (
                                        <VersionItem
                                            key={index}
                                            ver={ver}
                                            setVersion={setVersion}
                                        />
                                    );
                                })}
                            </div>
                        </>
                    )}
                </div>
            </div>
        </Modal>
    );
};

export default DetailModal;
