/*
 * Copyright by LunaSec (owned by Refinery Labs, Inc)
 *
 * Licensed under the Business Source License v1.1
 * (the "License"); you may not use this file except in compliance with the
 * License. You may obtain a copy of the License at
 *
 * https://github.com/lunasec-io/lunasec/blob/master/licenses/BSL-LunaTrace.txt
 *
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { TextInput } from '@ory/themes';
import React from 'react';

import { NodeInputProps } from './helpers';

export function NodeInputDefault<T>(props: NodeInputProps) {
  const { node, attributes, value = '', setValue, disabled } = props;

  // Some attributes have dynamic JavaScript - this is for example required for WebAuthn.
  const onClick = () => {
    // This section is only used for WebAuthn. The script is loaded via a <script> node
    // and the functions are available on the global window level. Unfortunately, there
    // is currently no better way than executing eval / function here at this moment.
    if (attributes.onclick) {
      // eslint-disable-next-line @typescript-eslint/no-implied-eval
      const run = new Function(attributes.onclick);
      run();
    }
  };

  const subtitle = (
    <>
      {node.messages.map(({ text, id }, k) => (
        <span key={`${id}-${k}`} data-testid={`ui/message/${id}`}>
          {text}
        </span>
      ))}
    </>
  );

  // Render a generic text input field.
  return (
    <TextInput
      title={node.meta.label?.text}
      className={'ory-input'}
      onClick={onClick}
      onChange={(e) => {
        void setValue(e.target.value);
      }}
      type={attributes.type}
      name={attributes.name}
      value={value}
      disabled={attributes.disabled || disabled}
      help={node.messages.length > 0}
      state={node.messages.find(({ type }) => type === 'error') ? 'error' : 'success'}
      subtitle={subtitle}
    />
  );
}
