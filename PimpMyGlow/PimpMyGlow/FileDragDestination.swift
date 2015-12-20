//
//  FileDragDestination.swift
//  PimpMyGlow
//
//  Created by Mark Probst on 12/19/15.
//  Copyright © 2015 Mark Probst. All rights reserved.
//

import Cocoa

protocol FileDragDestinationDelegate {
    func draggedFile(sender: FileDragDestination, path: String)
}

class FileDragDestination: NSView {
    var path: String?
    var delegate: FileDragDestinationDelegate?
    var mustBeDirectory: Bool
    
    override init(frame: NSRect) {
        self.mustBeDirectory = false
        super.init(frame: frame)
        let types = [NSFilenamesPboardType]
        registerForDraggedTypes(types)
    }

    required init?(coder: NSCoder) {
        self.mustBeDirectory = false
        super.init(coder: coder)
    }
    
    override func drawRect(dirtyRect: NSRect) {
        super.drawRect(dirtyRect)

        if path != nil {
            NSColor.greenColor().set()
        } else {
            NSColor.redColor().set()
        }
        NSRectFill(dirtyRect)
    }

    func draggingInfoCorrect(di: NSDraggingInfo) -> Bool {
        let pboard = di.draggingPasteboard()
        let files = pboard.propertyListForType(NSFilenamesPboardType)! as! [String]
        if files.count != 1 {
            return false
        }
        let file = files[0]
        var isDirectory: ObjCBool = false
        let exists = NSFileManager.defaultManager().fileExistsAtPath(file, isDirectory: &isDirectory)
        if !exists {
            return false
        }
        return isDirectory.boolValue == self.mustBeDirectory
    }
    
    override func draggingEntered(sender: NSDraggingInfo) -> NSDragOperation  {
        if !self.draggingInfoCorrect(sender) {
            return NSDragOperation.None
        }
        return NSDragOperation.Copy
    }
    
    override func prepareForDragOperation(sender: NSDraggingInfo) -> Bool {
        return self.draggingInfoCorrect(sender)
    }
    
    override func performDragOperation(sender: NSDraggingInfo) -> Bool {
        if !self.draggingInfoCorrect(sender) {
            return false
        }
        
        let pboard = sender.draggingPasteboard()
        // FIXME: check optional and cast
        let files = pboard.propertyListForType(NSFilenamesPboardType)! as! [String]
        self.path = files[0]
        self.setNeedsDisplayInRect(self.bounds)
        if self.delegate != nil {
            self.delegate!.draggedFile(self, path: self.path!)
        }
        return true
    }
}
