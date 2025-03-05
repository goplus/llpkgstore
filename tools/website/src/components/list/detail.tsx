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

const DetailModal: React.FC<DetailModalProps> = ({ modalOpen, setModalOpen, name, data }) => {
    const [originVersion, setOriginVersion] = useState('');
    const [convertVersion, setConvertVersion] = useState('');
    return (
        <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title={name}>
            {/* <Details /> */}
            {data && name && (
                <>
                    <div className="flex flex-row justify-around text-base my-2">
                        <div className="px-2 py-1 border border-gray-300 rounded-3xl flex items-center gap-2">
                            <input
                                onChange={(e) => setOriginVersion(e.target.value)}
                                className="focus-visible:outline-0 px-1"
                                placeholder="Original version"
                            />
                        </div>
                        <div className="px-2 py-1 border border-gray-300 rounded-3xl flex items-center gap-2">
                            <input
                                onChange={(e) => setConvertVersion(e.target.value)}
                                className="focus-visible:outline-0 px-1"
                                placeholder="Converted version"
                            />
                        </div>
                    </div>
                    {versionParser(data[name], originVersion, convertVersion, 1, 10).map((ver, index) => {
                        return (
                            <div key={index} className="flex flex-row gap-4 items-center">
                                <span className="font-bold text-lg min-w-16 text-left">{ver.original}</span>
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
